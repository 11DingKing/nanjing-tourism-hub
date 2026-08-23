package tourism

import (
	"context"
	"sort"
	"strings"
	"sync"
)

type Catalog struct {
	mu        sync.RWMutex
	Agencies  map[string]Agency
	Campaigns map[string]Campaign
	Routes    map[string]Route
	Packages  map[string]Package
}

func NewCatalog() *Catalog {
	return &Catalog{Agencies: map[string]Agency{}, Campaigns: map[string]Campaign{}, Routes: map[string]Route{}, Packages: map[string]Package{}}
}
func (c *Catalog) PutAgency(ctx context.Context, a Agency) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if a.ID == "" || strings.TrimSpace(a.Name) == "" {
		return ErrInvalidState
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if old, ok := c.Agencies[a.ID]; ok && old.Version != a.Version {
		return ErrConflict
	}
	if a.Version == 0 {
		a.Version = 1
	}
	c.Agencies[a.ID] = a
	return nil
}
func (c *Catalog) Agency(ctx context.Context, id string) (Agency, error) {
	if err := ctx.Err(); err != nil {
		return Agency{}, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	a, ok := c.Agencies[id]
	if !ok {
		return Agency{}, ErrNotFound
	}
	return a, nil
}
func (c *Catalog) ListAgencies(ctx context.Context, country string) ([]Agency, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := []Agency{}
	for _, a := range c.Agencies {
		if country == "" || strings.EqualFold(country, a.Country) {
			out = append(out, a)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}
