package sla

import (
	"testing"
	"time"
)

func TestSlaEscalation(t *testing.T) {
	m := New()
	opened := time.Now().Add(-2 * time.Hour)
	if e := m.Open(Case{ID: "c", RegionID: "r", Severity: Critical, OpenedAt: &opened}); e != nil {
		t.Fatal(e)
	}
	if len(m.Due(time.Now())) != 1 {
		t.Fatal("critical due")
	}
	if e := m.Respond("c", "u", time.Now()); e != nil {
		t.Fatal(e)
	}
	if e := m.Close("c", time.Now()); e != nil {
		t.Fatal(e)
	}
	if len(m.Due(time.Now().Add(24*time.Hour))) != 0 {
		t.Fatal("closed due")
	}
}
