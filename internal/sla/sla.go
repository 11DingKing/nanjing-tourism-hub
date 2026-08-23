package sla

import (
	"errors"
	"sort"
	"sync"
	"time"
)

type Severity string

const (
	Low      Severity = "low"
	Medium   Severity = "medium"
	High     Severity = "high"
	Critical Severity = "critical"
)

type Policy struct {
	Severity             Severity
	Response, Resolution time.Duration
	EscalationAfter      int
}
type Case struct {
	ID, RegionID, Owner                 string
	Severity                            Severity
	OpenedAt, FirstResponseAt, ClosedAt *time.Time
	Escalated                           bool
}
type Monitor struct {
	mu       sync.Mutex
	policies map[Severity]Policy
	cases    map[string]Case
}

func New() *Monitor {
	return &Monitor{policies: map[Severity]Policy{Low: {Low, 72 * time.Hour, 7 * 24 * time.Hour, 2}, Medium: {Medium, 24 * time.Hour, 72 * time.Hour, 1}, High: {High, 4 * time.Hour, 24 * time.Hour, 1}, Critical: {Critical, time.Hour, 8 * time.Hour, 1}}, cases: map[string]Case{}}
}
func (m *Monitor) Open(c Case) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c.ID == "" || c.RegionID == "" || c.OpenedAt == nil {
		return errors.New("case identity required")
	}
	if _, ok := m.cases[c.ID]; ok {
		return errors.New("case exists")
	}
	m.cases[c.ID] = c
	return nil
}
func (m *Monitor) Respond(id, owner string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.cases[id]
	if !ok {
		return errors.New("case not found")
	}
	if c.FirstResponseAt != nil {
		return errors.New("already responded")
	}
	c.Owner = owner
	c.FirstResponseAt = &at
	m.cases[id] = c
	return nil
}
func (m *Monitor) Close(id string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.cases[id]
	if !ok {
		return errors.New("case not found")
	}
	if c.FirstResponseAt == nil {
		return errors.New("case not responded")
	}
	c.ClosedAt = &at
	m.cases[id] = c
	return nil
}
func (m *Monitor) Due(now time.Time) []Case {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []Case{}
	for _, c := range m.cases {
		p := m.policies[c.Severity]
		if c.ClosedAt != nil {
			continue
		}
		if c.FirstResponseAt == nil && now.Sub(*c.OpenedAt) >= p.Response {
			c.Escalated = true
			out = append(out, c)
			continue
		}
		if c.FirstResponseAt != nil && now.Sub(*c.FirstResponseAt) >= p.Resolution {
			c.Escalated = true
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].OpenedAt.Before(*out[j].OpenedAt) })
	return out
}
func (m *Monitor) Policy(s Severity) Policy { return m.policies[s] }
