package tenant

import (
	"context"
	"testing"
)

func TestScopeIsolation(t *testing.T) {
	s, e := Parse("r1,r2,r1")
	if e != nil || len(s.RegionIDs) != 2 {
		t.Fatal(e)
	}
	if !s.Allows("r1") || s.Allows("r3") {
		t.Fatal("scope")
	}
	if len(s.Filter([]string{"r1", "r3", "r2"})) != 2 {
		t.Fatal("filter")
	}
	r := NewRegistry()
	if e = r.Set("u", s); e != nil {
		t.Fatal(e)
	}
	if e = r.Require(context.Background(), "u", "r3"); e == nil {
		t.Fatal("cross region")
	}
}
func TestScopeIntersection(t *testing.T) {
	a, _ := Parse("r1,r2")
	b, _ := Parse("r2,r3")
	c := Intersect(a, b)
	if len(c.RegionIDs) != 1 || !c.Allows("r2") {
		t.Fatal("intersection")
	}
	g, _ := Parse("*")
	if !Intersect(g, a).Allows("r1") {
		t.Fatal("global")
	}
}
