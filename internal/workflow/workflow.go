package workflow

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type State string

const (
	Pending   State = "pending"
	Running   State = "running"
	Succeeded State = "succeeded"
	Failed    State = "failed"
	Paused    State = "paused"
	Cancelled State = "cancelled"
)

type Step struct {
	ID         string
	Run        func(context.Context) error
	Compensate func(context.Context) error
	State      State
	Attempts   int
	Error      string
}
type Workflow struct {
	mu      sync.Mutex
	ID      string
	Steps   []Step
	State   State
	Current int
	History []string
	Version int
}

func New(id string, steps []Step) *Workflow {
	for i := range steps {
		if steps[i].State == "" {
			steps[i].State = Pending
		}
	}
	return &Workflow{ID: id, Steps: steps, State: Pending, Version: 1}
}
func (w *Workflow) Start(ctx context.Context) error {
	w.mu.Lock()
	if w.State != Pending && w.State != Paused {
		w.mu.Unlock()
		return errors.New("workflow cannot start")
	}
	w.State = Running
	w.Version++
	w.History = append(w.History, "started")
	w.mu.Unlock()
	for {
		w.mu.Lock()
		idx := w.Current
		if idx >= len(w.Steps) {
			w.State = Succeeded
			w.History = append(w.History, "succeeded")
			w.mu.Unlock()
			return nil
		}
		step := &w.Steps[idx]
		step.State = Running
		step.Attempts++
		w.mu.Unlock()
		e := step.Run(ctx)
		w.mu.Lock()
		if e != nil {
			step.State = Failed
			step.Error = e.Error()
			w.State = Failed
			w.History = append(w.History, fmt.Sprintf("step %s failed", step.ID))
			w.mu.Unlock()
			return e
		}
		step.State = Succeeded
		w.Current++
		w.History = append(w.History, fmt.Sprintf("step %s succeeded", step.ID))
		w.mu.Unlock()
	}
}
func (w *Workflow) Pause() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.State != Running {
		return errors.New("workflow not running")
	}
	w.State = Paused
	w.History = append(w.History, "paused")
	return nil
}
func (w *Workflow) Cancel(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.State == Succeeded || w.State == Cancelled {
		return nil
	}
	for i := w.Current - 1; i >= 0; i-- {
		if w.Steps[i].Compensate != nil {
			if e := w.Steps[i].Compensate(ctx); e != nil {
				return e
			}
		}
		w.Steps[i].State = Cancelled
	}
	w.State = Cancelled
	w.History = append(w.History, "cancelled")
	return nil
}
func (w *Workflow) Snapshot() ([]Step, State, int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	steps := append([]Step(nil), w.Steps...)
	return steps, w.State, w.Version
}
func (w *Workflow) Retryable() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.State == Failed && w.Current < len(w.Steps) && w.Steps[w.Current].Attempts < 3
}
func (w *Workflow) Retry(ctx context.Context) error {
	w.mu.Lock()
	if !(w.State == Failed && w.Current < len(w.Steps) && w.Steps[w.Current].Attempts < 3) {
		w.mu.Unlock()
		return errors.New("workflow not retryable")
	}
	w.State = Paused
	w.History = append(w.History, "retry scheduled")
	w.mu.Unlock()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Millisecond):
		return w.Start(ctx)
	}
}
