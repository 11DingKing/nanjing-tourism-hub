package tourism

import (
	"context"
	"sync"
	"time"
)

type Workflow struct {
	mu       sync.Mutex
	Steps    []string
	Done     []string
	Canceled bool
}

func (w *Workflow) Add(step string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if step != "" {
		w.Steps = append(w.Steps, step)
	}
}
func (w *Workflow) Run(ctx context.Context, fn func(context.Context, string) error) error {
	w.mu.Lock()
	steps := append([]string(nil), w.Steps...)
	w.mu.Unlock()
	for _, step := range steps {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := fn(ctx, step); err != nil {
			return err
		}
		w.mu.Lock()
		w.Done = append(w.Done, step)
		w.mu.Unlock()
	}
	return nil
}
func (w *Workflow) Cancel() { w.mu.Lock(); w.Canceled = true; w.mu.Unlock() }
func (w *Workflow) Deadline(now time.Time, hours int) time.Time {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return now.Add(time.Duration(hours) * time.Hour)
}
