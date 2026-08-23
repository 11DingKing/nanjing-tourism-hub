package allocation

import (
	"context"
	"testing"
	"time"
)

func TestPriorityAllocationAndRelease(t *testing.T) {
	p := New()
	if e := p.AddResource(Resource{ID: "dryer", RegionID: "r1", Kind: "drying", Capacity: 10, Available: 10}); e != nil {
		t.Fatal(e)
	}
	now := time.Now()
	if e := p.Submit(Request{ID: "low", RegionID: "r1", ResourceKind: "drying", OwnerID: "a", Amount: 7, Priority: 1, CreatedAt: now}); e != nil {
		t.Fatal(e)
	}
	if e := p.Submit(Request{ID: "high", RegionID: "r1", ResourceKind: "drying", OwnerID: "b", Amount: 7, Priority: 2, CreatedAt: now.Add(time.Second)}); e != nil {
		t.Fatal(e)
	}
	g, e := p.Allocate(context.Background(), now)
	if e != nil || g.RequestID != "high" {
		t.Fatalf("%+v %v", g, e)
	}
	if e = p.Confirm(g.ID); e != nil {
		t.Fatal(e)
	}
	r, _ := p.Resource("dryer")
	if r.Available != 3 {
		t.Fatal(r)
	}
}
func TestAllocationExpiry(t *testing.T) {
	p := New()
	_ = p.AddResource(Resource{ID: "r", RegionID: "x", Kind: "pump", Capacity: 2, Available: 2})
	_ = p.Submit(Request{ID: "q", RegionID: "x", ResourceKind: "pump", Amount: 2})
	g, e := p.Allocate(context.Background(), time.Now())
	if e != nil {
		t.Fatal(e)
	}
	if n := p.Expire(time.Now().Add(25 * time.Hour)); n != 1 {
		t.Fatal(n)
	}
	if e = p.Cancel(g.ID); e != nil {
		t.Fatal(e)
	}
	r, _ := p.Resource("r")
	if r.Available != 2 {
		t.Fatal(r)
	}
}
