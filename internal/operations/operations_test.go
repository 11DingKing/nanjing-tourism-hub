package operations

import (
	"context"
	"testing"
	"time"
)

func TestQueueOwnershipAndLease(t *testing.T) {
	q := NewQueue()
	now := time.Now()
	if e := q.Enqueue(Task{ID: "t1", RegionID: "r", PlotID: "p", Kind: Spray, CreatedAt: now, Metadata: map[string]string{"x": "y"}}); e != nil {
		t.Fatal(e)
	}
	t1, e := q.Claim("a", time.Minute)
	if e != nil || t1.Owner != "a" {
		t.Fatal(e)
	}
	if e := q.Finish("t1", "b", Succeeded, ""); e == nil {
		t.Fatal("wrong owner")
	}
	if e := q.Finish("t1", "a", Succeeded, ""); e != nil {
		t.Fatal(e)
	}
	got, _ := q.Get("t1")
	got.Metadata["x"] = "changed"
	if q.List("", Succeeded)[0].Metadata["x"] != "y" {
		t.Fatal("metadata alias")
	}
}
func TestRunnerRetriesAndContext(t *testing.T) {
	q := NewQueue()
	_ = q.Enqueue(Task{ID: "t", RegionID: "r", PlotID: "p", Kind: Inspection})
	r := Runner{Queue: q, MaxAttempts: 2, Handler: func(ctx context.Context, _ Task) error { return ctx.Err() }}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := r.RunOnce(ctx, "w"); e == nil {
		t.Fatal("cancel hidden")
	}
	if got, _ := q.Get("t"); got.Status != Failed && got.Status != Aborted {
		t.Fatalf("status=%s", got.Status)
	}
}
