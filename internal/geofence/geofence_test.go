package geofence

import "testing"

func TestPolygonContainment(t *testing.T) {
	f := New()
	if e := f.Add(Polygon{ID: "p", RegionID: "r", Vertices: []Point{{0, 0}, {0, 1}, {1, 1}, {1, 0}}}); e != nil {
		t.Fatal(e)
	}
	if !f.Contains("p", Point{.5, .5}) || f.Contains("p", Point{2, 2}) {
		t.Fatal("containment")
	}
	if _, ok := f.Nearest(Point{.6, .6}, "r"); !ok {
		t.Fatal("nearest")
	}
}
func TestPointValidation(t *testing.T) {
	if ValidPoint(Point{91, 0}) || !ValidPoint(Point{90, 180}) {
		t.Fatal("point validation")
	}
	if Distance(Point{0, 0}, Point{3, 4}) != 5 {
		t.Fatal("distance")
	}
}
