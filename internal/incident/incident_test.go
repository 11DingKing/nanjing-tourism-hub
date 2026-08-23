package incident

import (
	"context"
	"testing"
	"time"
)

func TestIncidentLifecycle(t *testing.T) {
	r := New()
	if e := r.Report(Incident{ID: "i", RegionID: "r", Reporter: "u", Kind: "flood", Description: "water over field", Severity: 4}); e != nil {
		t.Fatal(e)
	}
	for _, step := range []Status{Triaged, Assigned, Mitigating} {
		if e := r.Transition("i", step, "operator"); e != nil {
			t.Fatal(e)
		}
	}
	if e := r.AddAction(Action{ID: "a", IncidentID: "i", Actor: "crew", Kind: "drain"}); e != nil {
		t.Fatal(e)
	}
	if e := r.FinishAction("a", "complete"); e != nil {
		t.Fatal(e)
	}
	if e := r.Transition("i", Resolved, "operator"); e != nil {
		t.Fatal(e)
	}
	if len(r.List("r", Resolved)) != 1 {
		t.Fatal("resolved list")
	}
}
func TestIncidentValidationAndDue(t *testing.T) {
	r := New()
	if e := Validate(context.Background(), Incident{Severity: 9, Description: "long enough"}); e == nil {
		t.Fatal("severity accepted")
	}
	old := time.Now().Add(-time.Hour)
	_ = r.Report(Incident{ID: "i", RegionID: "r", Reporter: "u", Kind: "drought", Description: "crop stress", CreatedAt: old, UpdatedAt: old})
	if len(r.Due(time.Now(), time.Minute)) != 1 {
		t.Fatal("due")
	}
	i, _ := r.Get("i")
	i.Notes = append(i.Notes, "mutate")
	if len(r.GetNotesForTest("i")) != 0 {
		t.Fatal("copy")
	}
}
func TestIncidentLinksAndSeverity(t *testing.T) {
	r := New()
	_ = r.Report(Incident{ID: "a", RegionID: "r", Reporter: "u", Kind: "flood", Description: "long enough"})
	_ = r.Report(Incident{ID: "b", RegionID: "r", Reporter: "u", Kind: "rain", Description: "long enough"})
	if e := r.Link("a", "b"); e != nil {
		t.Fatal(e)
	}
	if e := r.Severity("a", 5); e != nil {
		t.Fatal(e)
	}
	a, _ := r.Get("a")
	if len(a.Related) != 1 || a.Severity != 5 {
		t.Fatal(a)
	}
	if e := r.Unlink("a", "b"); e != nil {
		t.Fatal(e)
	}
}
func (r *Registry) GetNotesForTest(id string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]string(nil), r.incidents[id].Notes...)
}
