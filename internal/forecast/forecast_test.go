package forecast

import (
	"testing"
	"time"
)

func TestEvaluateWarning(t *testing.T) {
	now := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)
	w := Evaluate("r1", []Reading{{RegionID: "r1", ObservedAt: now, RainMM: 120, Moisture: .8}}, now)
	if w.Level != "critical" || w.Code != "flood" {
		t.Fatalf("%+v", w)
	}
	if len(w.Evidence) != 1 {
		t.Fatal("evidence")
	}
}
func TestCatalogCopiesAndExpiry(t *testing.T) {
	now := time.Now()
	c := NewCatalog()
	w := Evaluate("r1", []Reading{{RegionID: "r1", ObservedAt: now, RainMM: 1}}, now)
	if e := c.Put(w); e != nil {
		t.Fatal(e)
	}
	got, ok := c.Get("r1")
	if !ok {
		t.Fatal("missing")
	}
	got.Evidence[0].RainMM = 999
	again, _ := c.Get("r1")
	if again.Evidence[0].RainMM == 999 {
		t.Fatal("evidence alias")
	}
	if len(c.Active(now.Add(7*time.Hour))) != 0 {
		t.Fatal("expired warning active")
	}
}
func TestForecastHelpers(t *testing.T) {
	now := time.Now()
	rs := []Reading{{RegionID: "r1", ObservedAt: now.Add(-time.Hour), Moisture: .2}, {RegionID: "r1", ObservedAt: now, RainMM: 4}, {RegionID: "r2", ObservedAt: now}}
	if Aggregate(rs).Moisture <= 0 {
		t.Fatal("aggregate")
	}
	if len(GroupByRegion(rs)) != 2 || Latest(rs)["r1"].RainMM != 4 {
		t.Fatal("group")
	}
	if !Stale(now, now.Add(-2*time.Hour), time.Hour) {
		t.Fatal("stale")
	}
}
