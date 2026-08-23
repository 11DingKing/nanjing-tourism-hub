package workflow

import (
	"context"
	"errors"
	"testing"
)

func TestWorkflowSuccessAndCompensation(t *testing.T) {
	ran := 0
	comp := 0
	w := New("w", []Step{{ID: "a", Run: func(context.Context) error { ran++; return nil }, Compensate: func(context.Context) error { comp++; return nil }}, {ID: "b", Run: func(context.Context) error { ran++; return nil }}})
	if e := w.Start(context.Background()); e != nil {
		t.Fatal(e)
	}
	if ran != 2 {
		t.Fatal(ran)
	}
	if e := w.Cancel(context.Background()); e != nil {
		t.Fatal(e)
	}
	if comp != 0 {
		t.Fatal("compensated succeeded workflow")
	}
}
func TestWorkflowFailureRetry(t *testing.T) {
	attempts := 0
	w := New("w", []Step{{ID: "a", Run: func(context.Context) error {
		attempts++
		if attempts < 2 {
			return errors.New("temporary")
		}
		return nil
	}}})
	if e := w.Start(context.Background()); e == nil {
		t.Fatal("failure hidden")
	}
	if !w.Retryable() {
		t.Fatal("not retryable")
	}
	if e := w.Retry(context.Background()); e != nil {
		t.Fatal(e)
	}
	_, state, _ := w.Snapshot()
	if state != Succeeded {
		t.Fatal(state)
	}
}
func TestWorkflowCancellationContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	w := New("w", []Step{{ID: "a", Run: func(ctx context.Context) error { return ctx.Err() }}})
	if e := w.Start(ctx); e == nil {
		t.Fatal("cancel hidden")
	}
	if e := w.Cancel(context.Background()); e != nil {
		t.Fatal(e)
	}
}
