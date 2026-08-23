package incident

import "testing"

func TestIncidentSummary(t *testing.T) {
	r := New()
	_ = r.Report(Incident{ID: "a", RegionID: "r", Reporter: "u", Kind: "flood", Description: "long enough", Severity: 4})
	_ = r.Report(Incident{ID: "b", RegionID: "r", Reporter: "u", Kind: "rain", Description: "long enough", Severity: 1})
	_ = r.Transition("a", Triaged, "x")
	s := r.Summary("r")
	if s.Total != 2 || s.Open != 2 || s.HighSeverity != 1 || s.Oldest == nil {
		t.Fatalf("%+v", s)
	}
}
