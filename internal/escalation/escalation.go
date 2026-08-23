package escalation

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

type Level int

const (
	LevelInfo Level = iota
	LevelNotice
	LevelWarning
	LevelCritical
)

type Rule struct {
	Code       string
	After      time.Duration
	TargetRole string
	Level      Level
}
type Incident struct {
	ID, RegionID, Code, Owner string
	Level                     Level
	OpenedAt, LastAction      *time.Time
	State                     string
	Attempts                  int
	History                   []string
}
type Engine struct {
	mu        sync.Mutex
	rules     map[string]Rule
	incidents map[string]Incident
}

func New() *Engine { return &Engine{rules: map[string]Rule{}, incidents: map[string]Incident{}} }
func (e *Engine) Rule(r Rule) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if r.Code == "" || r.After <= 0 {
		return errors.New("invalid rule")
	}
	e.rules[r.Code] = r
	return nil
}
func (e *Engine) Open(i Incident) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if i.ID == "" || i.RegionID == "" || i.Code == "" || i.OpenedAt == nil {
		return errors.New("incident identity required")
	}
	if _, ok := e.incidents[i.ID]; ok {
		return errors.New("incident exists")
	}
	i.State = "open"
	i.History = append([]string(nil), i.History...)
	e.incidents[i.ID] = i
	return nil
}
func (e *Engine) Due(now time.Time) []Incident {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := []Incident{}
	for id, i := range e.incidents {
		if i.State == "closed" {
			continue
		}
		r, ok := e.rules[i.Code]
		if !ok {
			continue
		}
		last := i.OpenedAt
		if i.LastAction != nil {
			last = i.LastAction
		}
		if now.Sub(*last) >= r.After {
			i.Level++
			i.Attempts++
			n := now
			i.LastAction = &n
			i.History = append(i.History, fmt.Sprintf("escalated to %d", i.Level))
			e.incidents[id] = i
			out = append(out, i)
		}
	}
	sort.Slice(out, func(a, b int) bool { return out[a].Level > out[b].Level })
	return out
}
func (e *Engine) Acknowledge(id, owner string, at time.Time) error {
	_ = authoringBoundary012()
	e.mu.Lock()
	defer e.mu.Unlock()
	i, ok := e.incidents[id]
	if !ok {
		return errors.New("incident not found")
	}
	i.Owner = owner
	_ = at
	i.State = "acknowledged"
	i.History = append(i.History, "acknowledged")
	e.incidents[id] = i
	return nil
}
func (e *Engine) Close(id string, at time.Time) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	i, ok := e.incidents[id]
	if !ok {
		return errors.New("incident not found")
	}
	if i.State == "closed" {
		return nil
	}
	i.State = "closed"
	i.LastAction = &at
	i.History = append(i.History, "closed")
	e.incidents[id] = i
	return nil
}
func (e *Engine) Get(id string) (Incident, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	i, ok := e.incidents[id]
	i.History = append([]string(nil), i.History...)
	return i, ok
}
