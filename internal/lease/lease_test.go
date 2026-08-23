package lease

import (
	"context"
	"testing"
	"time"
)

func TestLeaseOwnership(t *testing.T) {
	m := New()
	if _, e := m.Acquire(context.Background(), "task", "a", time.Minute); e != nil {
		t.Fatal(e)
	}
	if _, e := m.Acquire(context.Background(), "task", "b", time.Minute); e == nil {
		t.Fatal("double claim")
	}
	if e := m.Release(context.Background(), "task", "b"); e == nil {
		t.Fatal("wrong owner released")
	}
	if e := m.Release(context.Background(), "task", "a"); e != nil {
		t.Fatal(e)
	}
}
