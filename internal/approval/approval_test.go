package approval

import (
	"testing"
	"time"
)

func TestApprovalStateAndReviewerUniqueness(t *testing.T) {
	l := New()
	if e := l.Create(Item{ID: "a", RegionID: "r", Applicant: "farmer", AmountCents: 100, CreatedAt: time.Now()}); e != nil {
		t.Fatal(e)
	}
	for _, s := range []State{Submitted, Review} {
		if e := l.Transition("a", s, "worker", "ready"); e != nil {
			t.Fatal(e)
		}
	}
	v, _ := l.Get("a")
	v.Reviewers = []string{"u1", "u2"}
	l.mu.Lock()
	l.items["a"] = v
	l.mu.Unlock()
	if e := l.Decide("a", "u1", true, "duplicate"); e == nil {
		t.Fatal("duplicate reviewer")
	}
}
func TestApprovalPayment(t *testing.T) {
	l := New()
	_ = l.Create(Item{ID: "a", RegionID: "r", Applicant: "f", AmountCents: 100})
	_ = l.Transition("a", Submitted, "u", "submit")
	_ = l.Transition("a", Review, "u", "review")
	if e := l.Decide("a", "r1", true, "ok"); e != nil {
		t.Fatal(e)
	}
	if e := l.Transition("a", Paid, "finance", "paid"); e != nil {
		t.Fatal(e)
	}
}
func TestApprovalRejectsSkippingReviewToPayment(t *testing.T) {
	l := New()
	_ = l.Create(Item{ID: "a", RegionID: "r", Applicant: "f", AmountCents: 100})
	if e := l.Transition("a", Paid, "finance", "paid"); e == nil {
		t.Fatal("expected error transitioning draft to paid without review")
	}
	v, _ := l.Get("a")
	if v.State != Draft {
		t.Fatalf("state mutated to %s", v.State)
	}
}
