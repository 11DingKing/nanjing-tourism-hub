package health

import (
	"context"
	"github.com/11DingKing/autumn-grain-resilience/internal/storage"
	"testing"
)

func TestHealthDatabase(t *testing.T) {
	db, e := storage.Open(":memory:")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	s := Check(context.Background(), db)
	if e = Validate(s); e != nil {
		t.Fatal(e)
	}
	if e = Ready(context.Background(), db); e != nil {
		t.Fatal(e)
	}
}
func TestHealthNil(t *testing.T) {
	if e := Validate(Check(context.Background(), nil)); e == nil {
		t.Fatal("nil db healthy")
	}
}
func TestCompositeAndProbe(t *testing.T) {
	db, e := storage.Open(":memory:")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	s := Probe(context.Background(), db, 2)
	if !s.Healthy {
		t.Fatal(s)
	}
	if !DependencySummary([]Status{s}).Healthy {
		t.Fatal("summary")
	}
	if IsDegraded(s, 0) == false {
		t.Fatal("latency")
	}
}
