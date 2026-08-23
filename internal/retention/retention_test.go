package retention

import (
	"testing"
	"time"
)

func TestRetentionLegalHold(t *testing.T) {
	m := New()
	_ = m.Rule(Rule{Kind: "alert", Keep: time.Hour})
	created := time.Now().Add(-2 * time.Hour)
	_ = m.Add(Object{ID: "a", Kind: "alert", CreatedAt: created})
	if len(m.Eligible(time.Now())) != 1 {
		t.Fatal("eligible")
	}
	if e := m.Hold("a", true); e != nil {
		t.Fatal(e)
	}
	if e := m.Delete("a", time.Now()); e == nil {
		t.Fatal("hold ignored")
	}
	if e := m.Hold("a", false); e != nil {
		t.Fatal(e)
	}
	if e := m.Delete("a", time.Now()); e != nil {
		t.Fatal(e)
	}
	if m.CountActive() != 0 {
		t.Fatal("not deleted")
	}
}
