package lease

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Lease struct {
	ID, Owner string
	ExpiresAt time.Time
}
type Manager struct {
	mu    sync.Mutex
	items map[string]Lease
}

func New() *Manager { return &Manager{items: map[string]Lease{}} }
func (m *Manager) Acquire(_ context.Context, key, owner string, ttl time.Duration) (Lease, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	if old, ok := m.items[key]; ok && old.ExpiresAt.After(now) {
		return Lease{}, fmt.Errorf("lease held by %s", old.Owner)
	}
	v := Lease{ID: key, Owner: owner, ExpiresAt: now.Add(ttl)}
	m.items[key] = v
	return v, nil
}
func (m *Manager) Renew(_ context.Context, key, owner string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.items[key]
	if !ok || v.Owner != owner || time.Now().After(v.ExpiresAt) {
		return fmt.Errorf("lease unavailable")
	}
	v.ExpiresAt = time.Now().Add(ttl)
	m.items[key] = v
	return nil
}
func (m *Manager) Release(_ context.Context, key, owner string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.items[key]
	if !ok {
		return nil
	}
	if v.Owner != owner {
		return fmt.Errorf("lease owner mismatch")
	}
	delete(m.items, key)
	return nil
}
func (m *Manager) Snapshot() []Lease {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Lease, 0, len(m.items))
	for _, v := range m.items {
		out = append(out, v)
	}
	return out
}
