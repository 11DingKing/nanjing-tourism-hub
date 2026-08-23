package policy

import (
	"testing"
	"time"
)

func TestPolicyMatrix(t *testing.T) {
	cases := []struct {
		name string
		p    Principal
		r    Resource
		ok   bool
	}{{"admin", Principal{ID: "a", Role: "admin"}, Resource{RegionID: "r2"}, true}, {"same region", Principal{ID: "u", Role: "operator", Regions: map[string]bool{"r1": true}}, Resource{RegionID: "r1"}, true}, {"other region", Principal{ID: "u", Role: "operator", Regions: map[string]bool{"r1": true}}, Resource{RegionID: "r2"}, false}, {"empty", Principal{}, Resource{RegionID: "r1"}, false}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := CanMutate(tc.p, tc.r).Allowed; got != tc.ok {
				t.Fatalf("got %v want %v", got, tc.ok)
			}
		})
	}
}
func TestPolicyWindows(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if ValidateWindow(base, base).Error() == "" {
		t.Fatal("equal window accepted")
	}
	if ValidateWindow(base, base.Add(15*24*time.Hour)).Error() == "" {
		t.Fatal("long window accepted")
	}
	if ValidateWindow(base, base.Add(time.Hour)) != nil {
		t.Fatal("valid window rejected")
	}
}
func TestNormalizeFilter(t *testing.T) {
	got := NormalizeFilter(" High,low,high, ,LOW ")
	if len(got) != 2 || got[0] != "high" || got[1] != "low" {
		t.Fatalf("%v", got)
	}
}
