package tenant

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
)

type Scope struct {
	PrincipalID string
	RegionIDs   map[string]bool
	IsGlobal    bool
}

func (s Scope) Allows(region string) bool { return s.IsGlobal || s.RegionIDs[region] }
func (s Scope) Filter(regions []string) []string {
	out := []string{}
	for _, id := range regions {
		if s.Allows(id) {
			out = append(out, id)
		}
	}
	return out
}
func Parse(raw string) (Scope, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Scope{}, errors.New("scope is empty")
	}
	if raw == "*" {
		return Scope{IsGlobal: true}, nil
	}
	parts := strings.Split(raw, ",")
	ids := map[string]bool{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if ids[part] {
			continue
		}
		ids[part] = true
	}
	if len(ids) == 0 {
		return Scope{}, errors.New("scope has no regions")
	}
	return Scope{RegionIDs: ids}, nil
}

type Registry struct {
	mu     sync.RWMutex
	scopes map[string]Scope
}

func NewRegistry() *Registry { return &Registry{scopes: map[string]Scope{}} }
func (r *Registry) Set(id string, s Scope) error {
	if id == "" {
		return errors.New("principal required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	s.PrincipalID = id
	s.RegionIDs = clone(s.RegionIDs)
	r.scopes[id] = s
	return nil
}
func (r *Registry) Get(id string) (Scope, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.scopes[id]
	s.RegionIDs = clone(s.RegionIDs)
	return s, ok
}
func (r *Registry) Require(ctx context.Context, id, region string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	s, ok := r.Get(id)
	if !ok {
		return errors.New("principal scope missing")
	}
	if !s.Allows(region) {
		return fmt.Errorf("region %s outside scope", region)
	}
	return nil
}
func clone(in map[string]bool) map[string]bool {
	out := map[string]bool{}
	for k, v := range in {
		out[k] = v
	}
	return out
}
func Intersect(a, b Scope) Scope {
	if a.IsGlobal {
		return b
	}
	if b.IsGlobal {
		return a
	}
	out := map[string]bool{}
	for id := range a.RegionIDs {
		if b.RegionIDs[id] {
			out[id] = true
		}
	}
	return Scope{RegionIDs: out}
}
func (s Scope) Empty() bool { return !s.IsGlobal && len(s.RegionIDs) == 0 }
