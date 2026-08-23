package tourism

import (
	"context"
	"fmt"
	"time"
)

type VisitWorkflow struct {
	Name     string
	Attempts int
	LastAt   time.Time
}

func (w *VisitWorkflow) Describe() string { return "hosted visit check-in, confirmation and closure" }
func (w *VisitWorkflow) Validate(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if id == "" {
		return fmt.Errorf("%w: visits", ErrInvalidState)
	}
	return nil
}
func (w *VisitWorkflow) Execute(ctx context.Context, id string, now time.Time) error {
	if err := w.Validate(ctx, id); err != nil {
		return err
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	w.Attempts++
	w.LastAt = now
	return nil
}
func (w *VisitWorkflow) RetryAfter(now time.Time) time.Time {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	delay := time.Duration(w.Attempts*w.Attempts+1) * time.Second
	return now.Add(delay)
}
func (w *VisitWorkflow) Snapshot() map[string]any {
	return map[string]any{"name": w.Name, "attempts": w.Attempts, "last_at": w.LastAt}
}
