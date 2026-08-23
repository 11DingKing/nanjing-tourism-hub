package allocation

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

type Resource struct {
	ID, RegionID, Kind  string
	Capacity, Available float64
	Active              bool
	Version             int
}
type Request struct {
	ID, RegionID, ResourceKind, OwnerID string
	Amount                              float64
	Priority                            int
	CreatedAt                           time.Time
}
type Grant struct {
	ID, RequestID, ResourceID, OwnerID string
	Amount                             float64
	State                              string
	ExpiresAt                          time.Time
}
type Pool struct {
	mu        sync.Mutex
	resources map[string]Resource
	requests  map[string]Request
	grants    map[string]Grant
}

func New() *Pool {
	return &Pool{resources: map[string]Resource{}, requests: map[string]Request{}, grants: map[string]Grant{}}
}
func (p *Pool) AddResource(r Resource) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if r.ID == "" || r.RegionID == "" || r.Kind == "" {
		return errors.New("resource identity required")
	}
	if r.Capacity <= 0 || r.Available < 0 || r.Available > r.Capacity {
		return errors.New("invalid capacity")
	}
	if _, ok := p.resources[r.ID]; ok {
		return errors.New("resource exists")
	}
	r.Active = true
	if r.Version == 0 {
		r.Version = 1
	}
	p.resources[r.ID] = r
	return nil
}
func (p *Pool) Submit(r Request) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if r.ID == "" || r.RegionID == "" || r.ResourceKind == "" {
		return errors.New("request identity required")
	}
	if r.Amount <= 0 {
		return errors.New("amount must be positive")
	}
	if _, ok := p.requests[r.ID]; ok {
		return errors.New("request exists")
	}
	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now().UTC()
	}
	p.requests[r.ID] = r
	return nil
}
func (p *Pool) Allocate(_ context.Context, now time.Time) (Grant, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	reqs := make([]Request, 0, len(p.requests))
	for _, r := range p.requests {
		already := false
		for _, g := range p.grants {
			if g.RequestID == r.ID && g.State != "cancelled" {
				already = true
			}
		}
		if !already {
			reqs = append(reqs, r)
		}
	}
	sort.Slice(reqs, func(i, j int) bool {
		if reqs[i].Priority == reqs[j].Priority {
			return reqs[i].CreatedAt.Before(reqs[j].CreatedAt)
		}
		return reqs[i].Priority > reqs[j].Priority
	})
	for _, r := range reqs {
		for id, res := range p.resources {
			if !res.Active || res.RegionID != r.RegionID || res.Kind != r.ResourceKind || res.Available < r.Amount {
				continue
			}
			res.Available -= r.Amount
			res.Version++
			p.resources[id] = res
			g := Grant{ID: fmt.Sprintf("grant-%d", len(p.grants)+1), RequestID: r.ID, ResourceID: id, OwnerID: r.OwnerID, Amount: r.Amount, State: "held", ExpiresAt: now.Add(24 * time.Hour)}
			p.grants[g.ID] = g
			return g, nil
		}
	}
	return Grant{}, errors.New("resource unavailable")
}
func (p *Pool) Confirm(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	g, ok := p.grants[id]
	if !ok {
		return errors.New("grant not found")
	}
	if g.State != "held" {
		return errors.New("grant not held")
	}
	g.State = "confirmed"
	p.grants[id] = g
	return nil
}
func (p *Pool) Cancel(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	g, ok := p.grants[id]
	if !ok {
		return errors.New("grant not found")
	}
	if g.State == "cancelled" {
		return nil
	}
	r := p.resources[g.ResourceID]
	r.Available += g.Amount
	r.Version++
	p.resources[r.ID] = r
	g.State = "cancelled"
	p.grants[id] = g
	return nil
}
func (p *Pool) Expire(now time.Time) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	n := 0
	for id, g := range p.grants {
		if g.State == "held" && now.After(g.ExpiresAt) {
			r := p.resources[g.ResourceID]
			r.Available += g.Amount
			r.Version++
			p.resources[r.ID] = r
			g.State = "cancelled"
			p.grants[id] = g
			n++
		}
	}
	return n
}
func (p *Pool) Resource(id string) (Resource, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	r, ok := p.resources[id]
	return r, ok
}
func (p *Pool) Grants() []Grant {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]Grant, 0, len(p.grants))
	for _, g := range p.grants {
		out = append(out, g)
	}
	return out
}
