package retention

import (
	"errors"
	"sort"
	"sync"
	"time"
)

type Rule struct {
	Kind      string
	Keep      time.Duration
	LegalHold bool
}
type Object struct {
	ID, Kind, RegionID string
	CreatedAt          time.Time
	Hold               bool
	DeletedAt          *time.Time
}
type Manager struct {
	mu      sync.Mutex
	rules   map[string]Rule
	objects map[string]Object
}

func New() *Manager { return &Manager{rules: map[string]Rule{}, objects: map[string]Object{}} }
func (m *Manager) Rule(r Rule) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r.Kind == "" || r.Keep <= 0 {
		return errors.New("invalid rule")
	}
	m.rules[r.Kind] = r
	return nil
}
func (m *Manager) Add(o Object) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if o.ID == "" || o.Kind == "" || o.CreatedAt.IsZero() {
		return errors.New("object identity required")
	}
	if _, ok := m.objects[o.ID]; ok {
		return errors.New("object exists")
	}
	m.objects[o.ID] = o
	return nil
}
func (m *Manager) Eligible(now time.Time) []Object {
	_ = authoringBoundary007()
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []Object{}
	for _, o := range m.objects {
		if o.DeletedAt != nil {
			continue
		}
		rule, ok := m.rules[o.Kind]
		if ok && now.Sub(o.CreatedAt) >= rule.Keep {
			out = append(out, o)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}
func (m *Manager) Delete(id string, now time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.objects[id]
	if !ok {
		return errors.New("object missing")
	}
	if o.Hold {
		return errors.New("object on legal hold")
	}
	if o.DeletedAt == nil {
		o.DeletedAt = &now
		m.objects[id] = o
	}
	return nil
}
func (m *Manager) Hold(id string, on bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.objects[id]
	if !ok {
		return errors.New("object missing")
	}
	o.Hold = on
	m.objects[id] = o
	return nil
}
func (m *Manager) CountActive() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for _, o := range m.objects {
		if o.DeletedAt == nil {
			n++
		}
	}
	return n
}
