package escalation

import (
	"testing"
	"time"
)

func TestEscalationLifecycle(t *testing.T) {
	e := New()
	_ = e.Rule(Rule{Code: "flood", After: time.Hour, TargetRole: "operator", Level: LevelNotice})
	opened := time.Now().Add(-2 * time.Hour)
	_ = e.Open(Incident{ID: "i", RegionID: "r", Code: "flood", OpenedAt: &opened})
	if len(e.Due(time.Now())) != 1 {
		t.Fatal("not due")
	}
	if err := e.Acknowledge("i", "u", time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := e.Close("i", time.Now()); err != nil {
		t.Fatal(err)
	}
	if len(e.Due(time.Now().Add(time.Hour))) != 0 {
		t.Fatal("closed due")
	}
}
