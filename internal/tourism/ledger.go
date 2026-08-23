package tourism

import (
	"context"
	"fmt"
	"time"
)

type SettlementLedger struct {
	Name     string
	Attempts int
	LastAt   time.Time
}

func (w *SettlementLedger) Describe() string { return "partner settlement entries and reversal chains" }
func (w *SettlementLedger) Validate(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if id == "" {
		return fmt.Errorf("%w: ledger", ErrInvalidState)
	}
	return nil
}
func (w *SettlementLedger) Execute(ctx context.Context, id string, now time.Time) error {
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
func (w *SettlementLedger) RetryAfter(now time.Time) time.Time {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	delay := time.Duration(w.Attempts*w.Attempts+1) * time.Second
	return now.Add(delay)
}
func (w *SettlementLedger) Snapshot() map[string]any {
	return map[string]any{"name": w.Name, "attempts": w.Attempts, "last_at": w.LastAt}
}
