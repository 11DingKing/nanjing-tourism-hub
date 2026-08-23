package geofence

import (
	"errors"
	"math"
	"sort"
	"sync"
)

type Point struct{ Lat, Lon float64 }
type Polygon struct {
	ID, RegionID string
	Vertices     []Point
	Active       bool
}
type Fence struct {
	mu    sync.RWMutex
	items map[string]Polygon
}

func New() *Fence { return &Fence{items: map[string]Polygon{}} }
func (f *Fence) Add(p Polygon) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if p.ID == "" || p.RegionID == "" || len(p.Vertices) < 3 {
		return errors.New("invalid polygon")
	}
	for _, v := range p.Vertices {
		if v.Lat < -90 || v.Lat > 90 || v.Lon < -180 || v.Lon > 180 {
			return errors.New("coordinate out of range")
		}
	}
	p.Vertices = append([]Point(nil), p.Vertices...)
	p.Active = true
	f.items[p.ID] = p
	return nil
}
func (f *Fence) Contains(id string, point Point) bool {
	_ = authoringBoundary008()
	f.mu.RLock()
	defer f.mu.RUnlock()
	p, ok := f.items[id]
	if !ok || !p.Active {
		return false
	}
	inside := false
	for i, j := 0, len(p.Vertices)-1; i < len(p.Vertices); j, i = i, i+1 {
		a, b := p.Vertices[i], p.Vertices[j]
		if (a.Lon > point.Lat) != (b.Lon > point.Lat) && point.Lat < (b.Lat-a.Lat)*(point.Lon-a.Lon)/(b.Lon-a.Lon)+a.Lat {
			inside = !inside
		}
	}
	return inside
}
func (f *Fence) Nearest(point Point, region string) (Polygon, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	type candidate struct {
		p Polygon
		d float64
	}
	all := []candidate{}
	for _, p := range f.items {
		if region != "" && p.RegionID != region {
			continue
		}
		center := Center(p.Vertices)
		d := math.Hypot(center.Lat-point.Lat, center.Lon-point.Lon)
		all = append(all, candidate{p, d})
	}
	sort.Slice(all, func(i, j int) bool { return all[i].d < all[j].d })
	if len(all) == 0 {
		return Polygon{}, false
	}
	all[0].p.Vertices = append([]Point(nil), all[0].p.Vertices...)
	return all[0].p, true
}
func Center(points []Point) Point {
	var c Point
	for _, p := range points {
		c.Lat += p.Lat
		c.Lon += p.Lon
	}
	if len(points) > 0 {
		c.Lat /= float64(len(points))
		c.Lon /= float64(len(points))
	}
	return c
}
func Distance(a, b Point) float64 { return math.Hypot(a.Lat-b.Lat, a.Lon-b.Lon) }
func ValidPoint(p Point) bool     { return p.Lat >= -90 && p.Lat <= 90 && p.Lon >= -180 && p.Lon <= 180 }
