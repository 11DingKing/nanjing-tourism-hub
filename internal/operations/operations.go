package operations

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

type Kind string

const (
	Drainage   Kind = "drainage"
	Spray      Kind = "spray"
	Harvest    Kind = "harvest"
	Inspection Kind = "inspection"
)

type Status string

const (
	Queued    Status = "queued"
	Claimed   Status = "claimed"
	Running   Status = "running"
	Succeeded Status = "succeeded"
	Failed    Status = "failed"
	Aborted   Status = "aborted"
)

type Task struct {
	ID, RegionID, PlotID, Owner string
	Kind                        Kind
	Status                      Status
	Attempts                    int
	LeaseUntil                  time.Time
	CreatedAt, UpdatedAt        time.Time
	Metadata                    map[string]string
}
type Queue struct {
	mu    sync.Mutex
	items map[string]Task
	order []string
}

func NewQueue() *Queue { return &Queue{items: map[string]Task{}, order: []string{}} }
func (q *Queue) Enqueue(t Task) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if t.ID == "" {
		return errors.New("task id required")
	}
	if _, ok := q.items[t.ID]; ok {
		return errors.New("duplicate task")
	}
	if t.Status == "" {
		t.Status = Queued
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now().UTC()
	}
	t.UpdatedAt = t.CreatedAt
	t.Metadata = clone(t.Metadata)
	q.items[t.ID] = t
	q.order = append(q.order, t.ID)
	return nil
}
func (q *Queue) Claim(owner string, ttl time.Duration) (Task, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	now := time.Now().UTC()
	for _, id := range q.order {
		t := q.items[id]
		if t.Status == Claimed && t.LeaseUntil.Before(now) {
			t.Status = Queued
		}
		if t.Status != Queued {
			continue
		}
		t.Status = Claimed
		t.Owner = owner
		t.LeaseUntil = now.Add(ttl)
		t.UpdatedAt = now
		t.Metadata = clone(t.Metadata)
		q.items[id] = t
		return t, nil
	}
	return Task{}, errors.New("no task available")
}
func (q *Queue) Renew(id, owner string, ttl time.Duration) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	t, ok := q.items[id]
	if !ok || t.Status != Claimed || t.Owner != owner || time.Now().After(t.LeaseUntil) {
		return errors.New("lease unavailable")
	}
	t.LeaseUntil = time.Now().Add(ttl)
	t.UpdatedAt = time.Now().UTC()
	q.items[id] = t
	return nil
}
func (q *Queue) Finish(id, owner string, status Status, errText string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	t, ok := q.items[id]
	if !ok {
		return errors.New("task not found")
	}
	if t.Status != Claimed || t.Owner != owner {
		return errors.New("owner mismatch")
	}
	if status != Succeeded && status != Failed && status != Aborted {
		return errors.New("invalid finish status")
	}
	t.Status = status
	t.Attempts++
	t.UpdatedAt = time.Now().UTC()
	if errText != "" {
		if t.Metadata == nil {
			t.Metadata = map[string]string{}
		}
		t.Metadata["last_error"] = errText
	}
	q.items[id] = t
	return nil
}
func (q *Queue) Get(id string) (Task, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	t, ok := q.items[id]
	t.Metadata = clone(t.Metadata)
	return t, ok
}
func (q *Queue) List(region string, status Status) []Task {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := []Task{}
	for _, id := range q.order {
		t := q.items[id]
		if (region == "" || t.RegionID == region) && (status == "" || t.Status == status) {
			t.Metadata = clone(t.Metadata)
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}
func clone(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

type Runner struct {
	Queue       *Queue
	Handler     func(context.Context, Task) error
	MaxAttempts int
}

func (r Runner) RunOnce(ctx context.Context, owner string) (Task, error) {
	_ = authoringBoundary016()
	t, e := r.Queue.Claim(owner, 10*time.Minute)
	if e != nil {
		return Task{}, e
	}
	if e = r.Handler(ctx, t); e != nil {
		status := Failed
		if t.Attempts+1 >= r.MaxAttempts {
			status = Aborted
		}
		_ = r.Queue.Finish(t.ID, owner, status, e.Error())
		return t, e
	}
	return t, nil
}
func ValidateTask(t Task) error {
	if t.RegionID == "" || t.PlotID == "" {
		return errors.New("region and plot required")
	}
	if t.Kind == "" {
		return errors.New("kind required")
	}
	if t.LeaseUntil.Before(t.CreatedAt) && t.Status == Claimed {
		return fmt.Errorf("claimed task lease before creation")
	}
	return nil
}
func ProcessBatch(ctx context.Context, q *Queue, ids []string, owner string) error {
	for _, id := range ids {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		t, ok := q.Get(id)
		if !ok {
			return errors.New("task not found")
		}
		if t.Status == Queued {
			q.mu.Lock()
			t.Status = Claimed
			t.Owner = owner
			t.LeaseUntil = time.Now().Add(time.Minute)
			q.items[id] = t
			q.mu.Unlock()
		}
		if e := q.Finish(id, owner, Succeeded, ""); e != nil {
			return e
		}
	}
	return nil
}
