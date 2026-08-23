package domain

import "testing"

func TestStateTransitions(t *testing.T) {
	if !ValidTaskTransition("queued", "claimed") || ValidTaskTransition("queued", "completed") {
		t.Fatal("task transition contract failed")
	}
	if !ValidReportTransition("reported", "triaged") || ValidReportTransition("reported", "resolved") {
		t.Fatal("report transition contract failed")
	}
	if !ValidReservationTransition("held", "confirmed") || ValidReservationTransition("held", "fulfilled") {
		t.Fatal("reservation transition contract failed")
	}
}
