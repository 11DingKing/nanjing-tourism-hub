package integration

import (
	"context"
	"github.com/11DingKing/autumn-grain-resilience/internal/allocation"
	"github.com/11DingKing/autumn-grain-resilience/internal/approval"
	"github.com/11DingKing/autumn-grain-resilience/internal/cache"
	"github.com/11DingKing/autumn-grain-resilience/internal/compliance"
	"github.com/11DingKing/autumn-grain-resilience/internal/event"
	"github.com/11DingKing/autumn-grain-resilience/internal/forecast"
	"github.com/11DingKing/autumn-grain-resilience/internal/logistics"
	"github.com/11DingKing/autumn-grain-resilience/internal/notifications"
	"github.com/11DingKing/autumn-grain-resilience/internal/recovery"
	"github.com/11DingKing/autumn-grain-resilience/internal/tenant"
	"testing"
	"time"
)

func TestCrossModuleResilienceContract(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	pool := allocation.New()
	if e := pool.AddResource(allocation.Resource{ID: "dryer", RegionID: "r", Kind: "drying", Capacity: 20, Available: 20}); e != nil {
		t.Fatal(e)
	}
	if e := pool.Submit(allocation.Request{ID: "req", RegionID: "r", ResourceKind: "drying", OwnerID: "u", Amount: 10, Priority: 5, CreatedAt: now}); e != nil {
		t.Fatal(e)
	}
	if _, e := pool.Allocate(ctx, now); e != nil {
		t.Fatal(e)
	}
	ledger := approval.New()
	if e := ledger.Create(approval.Item{ID: "aid", RegionID: "r", Applicant: "farmer", AmountCents: 500}); e != nil {
		t.Fatal(e)
	}
	for _, s := range []approval.State{approval.Submitted, approval.Review} {
		if e := ledger.Transition("aid", s, "worker", "step"); e != nil {
			t.Fatal(e)
		}
	}
	if e := ledger.Decide("aid", "reviewer", true, "eligible"); e != nil {
		t.Fatal(e)
	}
	c := cache.New()
	if e := c.Set("warning", "high", time.Hour, 1); e != nil {
		t.Fatal(e)
	}
	if _, _, ok := c.Get("warning"); !ok {
		t.Fatal("cache")
	}
	reg := compliance.New()
	_ = reg.AddControl(compliance.Control{ID: "control", Name: "field", EvidenceKinds: []string{"photo"}})
	if e := reg.Collect(compliance.Evidence{ID: "ev", ControlID: "control", ObjectID: "plot", ActorID: "u", Kind: "photo"}, []byte("evidence")); e != nil {
		t.Fatal(e)
	}
	bus := event.New()
	seen := 0
	_ = bus.Subscribe("alert", func(_ context.Context, _ event.Event) error { seen++; return nil })
	if e := bus.Publish(ctx, event.Event{ID: "event", Topic: "alert"}); e != nil || seen != 1 {
		t.Fatalf("event %d %v", seen, e)
	}
	w := forecast.Evaluate("r", []forecast.Reading{{RegionID: "r", ObservedAt: now, RainMM: 110}}, now)
	if w.Level != "critical" {
		t.Fatal(w)
	}
	store := logistics.New()
	_ = store.Add(logistics.Item{ID: "seed", RegionID: "r", Name: "seed", Unit: "kg", Quantity: 100})
	if e := store.Move(logistics.Movement{ItemID: "seed", Kind: "reserve", Quantity: 20}); e != nil {
		t.Fatal(e)
	}
	sender := &notifications.MemorySender{}
	d := notifications.NewDispatcher()
	d.Senders[notifications.SMS] = sender
	if e := d.Dispatch(ctx, notifications.Message{ID: "m", Recipient: "farmer", Body: "warning", Channel: notifications.SMS}); e != nil {
		t.Fatal(e)
	}
	plan := recovery.NewPlan("plan", "r", "report", []recovery.Step{{ID: "inspect", State: recovery.Pending, DueAt: now.Add(time.Hour)}})
	if e := plan.Assign("inspect", "crew"); e != nil {
		t.Fatal(e)
	}
	if e := recovery.Run(ctx, plan, "inspect"); e != nil {
		t.Fatal(e)
	}
	if !plan.Complete() {
		t.Fatal("plan")
	}
	scope, _ := tenant.Parse("r")
	if !scope.Allows("r") || scope.Allows("other") {
		t.Fatal("tenant")
	}
}
