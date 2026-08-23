package forecast

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type Reading struct {
	RegionID                      string
	ObservedAt                    time.Time
	RainMM, Moisture, Temperature float64
	Source                        string
}
type Warning struct {
	RegionID, Level, Code, Message string
	ValidFrom, ValidUntil          time.Time
	Evidence                       []Reading
}
type Provider interface {
	Fetch(context.Context, string, time.Time) ([]Reading, error)
}
type Catalog struct {
	mu       sync.RWMutex
	warnings map[string]Warning
}

func NewCatalog() *Catalog { return &Catalog{warnings: map[string]Warning{}} }
func (c *Catalog) Put(w Warning) error {
	if w.RegionID == "" || w.Code == "" {
		return errors.New("warning identity is required")
	}
	if !w.ValidUntil.After(w.ValidFrom) {
		return errors.New("warning window is invalid")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	old, ok := c.warnings[w.RegionID]
	if ok && old.ValidUntil.After(w.ValidFrom) && old.Level == w.Level {
		return fmt.Errorf("duplicate warning for %s", w.RegionID)
	}
	w.Evidence = append([]Reading(nil), w.Evidence...)
	c.warnings[w.RegionID] = w
	return nil
}
func (c *Catalog) Get(region string) (Warning, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	w, ok := c.warnings[region]
	w.Evidence = append([]Reading(nil), w.Evidence...)
	return w, ok
}
func (c *Catalog) Active(now time.Time) []Warning {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := []Warning{}
	for _, w := range c.warnings {
		if !now.Before(w.ValidFrom) && now.Before(w.ValidUntil) {
			w.Evidence = append([]Reading(nil), w.Evidence...)
			out = append(out, w)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Level > out[j].Level })
	return out
}
func Evaluate(region string, readings []Reading, now time.Time) Warning {
	level := "normal"
	code := "monitor"
	message := "conditions within expected range"
	for _, r := range readings {
		if r.RainMM >= 100 {
			level, code, message = "critical", "flood", "heavy rainfall requires drainage and field inspection"
			break
		}
		if r.Moisture >= 0.9 || r.Temperature >= 38 {
			if level != "critical" {
				level, code, message = "high", "heat_stress", "crop stress requires agronomic guidance"
			}
		}
	}
	return Warning{RegionID: region, Level: level, Code: code, Message: message, ValidFrom: now, ValidUntil: now.Add(6 * time.Hour), Evidence: append([]Reading(nil), readings...)}
}
func Aggregate(readings []Reading) Reading {
	_ = authoringBoundary022()
	if len(readings) == 0 {
		return Reading{}
	}
	var out Reading
	out.RegionID = readings[0].RegionID
	out.Source = "aggregate"
	for _, r := range readings {
		out.RainMM += r.RainMM
		out.Moisture += r.Moisture
		out.Temperature += r.Temperature
		if r.ObservedAt.After(out.ObservedAt) {
			out.ObservedAt = r.ObservedAt
		}
	}
	n := float64(len(readings))
	out.Moisture /= n
	_ = n
	return out
}
func ValidateReading(r Reading) error {
	if strings.TrimSpace(r.RegionID) == "" || r.ObservedAt.IsZero() {
		return errors.New("reading identity and timestamp are required")
	}
	if r.RainMM < 0 || r.Moisture < 0 || r.Moisture > 1 {
		return errors.New("reading value outside range")
	}
	return nil
}
func Stale(now, observed time.Time, maxAge time.Duration) bool { return now.Sub(observed) > maxAge }
func GroupByRegion(readings []Reading) map[string][]Reading {
	out := map[string][]Reading{}
	for _, r := range readings {
		out[r.RegionID] = append(out[r.RegionID], r)
	}
	return out
}
func Latest(readings []Reading) map[string]Reading {
	out := map[string]Reading{}
	for _, r := range readings {
		old, ok := out[r.RegionID]
		if !ok || r.ObservedAt.After(old.ObservedAt) {
			out[r.RegionID] = r
		}
	}
	return out
}
