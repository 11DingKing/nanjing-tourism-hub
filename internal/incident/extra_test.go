package incident

import (
	"context"
	"testing"
	"time"
)

func TestIncidentInvalidTransitions(t *testing.T) {
	r := New()
	_ = r.Report(Incident{ID: "i", RegionID: "r", Reporter: "u", Kind: "flood", Description: "long enough"})
	for _, tc := range []struct {
		to   Status
		want bool
	}{{Assigned, false}, {Reopened, false}, {Triaged, true}} {
		e := r.Transition("i", tc.to, "actor")
		if (e == nil) != tc.want {
			t.Fatalf("to=%s err=%v", tc.to, e)
		}
	}
	if e := r.Transition("i", Triaged, " "); e == nil {
		t.Fatal("blank actor")
	}
}
func TestIncidentActionErrors(t *testing.T) {
	r := New()
	_ = r.Report(Incident{ID: "i", RegionID: "r", Reporter: "u", Kind: "flood", Description: "long enough"})
	if e := r.AddAction(Action{ID: "a", IncidentID: "i", Actor: "u", Kind: "x"}); e == nil {
		t.Fatal("action before mitigate")
	}
	_ = r.Transition("i", Triaged, "u")
	_ = r.Transition("i", Assigned, "u")
	_ = r.Transition("i", Mitigating, "u")
	if e := r.AddAction(Action{ID: "a", IncidentID: "i", Actor: "u", Kind: "x"}); e != nil {
		t.Fatal(e)
	}
	if e := r.AddAction(Action{ID: "a", IncidentID: "i", Actor: "u", Kind: "x"}); e == nil {
		t.Fatal("duplicate action")
	}
	if e := r.FinishAction("a", "ok"); e != nil {
		t.Fatal(e)
	}
	if e := r.FinishAction("a", "again"); e == nil {
		t.Fatal("finished action")
	}
	if len(r.ActionList("i")) != 1 {
		t.Fatal("action list")
	}
}
func TestIncidentContextValidation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if e := Validate(ctx, Incident{Severity: 1, Description: "long enough"}); e == nil {
		t.Fatal("cancel ignored")
	}
	if e := Validate(context.Background(), Incident{Severity: 1, Description: "x"}); e == nil {
		t.Fatal("short description")
	}
	if e := Validate(context.Background(), Incident{Severity: 1, Description: "long enough"}); e != nil {
		t.Fatal(e)
	}
}
func TestIncidentReopenAndOwner(t *testing.T) {
	r := New()
	_ = r.Report(Incident{ID: "i", RegionID: "r", Reporter: "u", Kind: "flood", Description: "long enough"})
	_ = r.Transition("i", Triaged, "u")
	_ = r.Transition("i", Assigned, "u")
	_ = r.Transition("i", Mitigating, "u")
	_ = r.Transition("i", Resolved, "u")
	if e := r.Reopen("i", "u"); e != nil {
		t.Fatal(e)
	}
	if e := r.SetOwner("i", "new"); e != nil {
		t.Fatal(e)
	}
	i, _ := r.Get("i")
	if i.Owner != "new" || i.Status != Reopened {
		t.Fatal(i)
	}
	if e := r.SetOwner("i", ""); e == nil {
		t.Fatal("blank owner")
	}
	_ = r.RemoveAction("missing")
	_ = r.Notes("i")
	_ = time.Now()
}
