package timeseries

import (
	"errors"
	"sort"
	"sync"
	"time"
)

type Point struct {
	At      time.Time
	Value   float64
	Quality string
}
type Series struct {
	ID, RegionID, Metric string
	Points               []Point
}
type Store struct {
	mu     sync.RWMutex
	series map[string]Series
}

func New() *Store { return &Store{series: map[string]Series{}} }
func (s *Store) Append(id, region, metric string, p Point) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id == "" || region == "" || metric == "" || p.At.IsZero() {
		return errors.New("series identity required")
	}
	v := s.series[id]
	if v.ID == "" {
		v = Series{ID: id, RegionID: region, Metric: metric}
	}
	if v.RegionID != region || v.Metric != metric {
		return errors.New("series mismatch")
	}
	if len(v.Points) > 0 && !p.At.After(v.Points[len(v.Points)-1].At) {
		return errors.New("point must be chronological")
	}
	v.Points = append(v.Points, p)
	s.series[id] = v
	return nil
}
func (s *Store) Window(id string, start, end time.Time) []Point {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v := s.series[id]
	out := []Point{}
	for _, p := range v.Points {
		if !p.At.Before(start) && p.At.Before(end) {
			out = append(out, p)
		}
	}
	return out
}
func (s *Store) Average(id string, start, end time.Time) (float64, error) {
	points := s.Window(id, start, end)
	if len(points) == 0 {
		return 0, errors.New("no points")
	}
	var total float64
	for _, p := range points {
		total += p.Value
	}
	return total / float64(len(points)), nil
}
func (s *Store) Latest(id string) (Point, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v := s.series[id]
	if len(v.Points) == 0 {
		return Point{}, false
	}
	return v.Points[len(v.Points)-1], true
}
func (s *Store) Series(region, metric string) []Series {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Series{}
	for _, v := range s.series {
		if (region == "" || v.RegionID == region) && (metric == "" || v.Metric == metric) {
			v.Points = append([]Point(nil), v.Points...)
			sort.Slice(v.Points, func(i, j int) bool { return v.Points[i].At.Before(v.Points[j].At) })
			out = append(out, v)
		}
	}
	return out
}
