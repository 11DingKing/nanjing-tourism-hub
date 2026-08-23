package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Entry struct {
	Value     any
	ExpiresAt time.Time
	Version   int
}
type Cache struct {
	mu    sync.RWMutex
	items map[string]Entry
	clock func() time.Time
}

func New() *Cache { return &Cache{items: map[string]Entry{}, clock: time.Now} }
func (c *Cache) Set(key string, value any, ttl time.Duration, version int) error {
	if key == "" {
		return errors.New("cache key required")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = Entry{Value: value, ExpiresAt: c.clock().Add(ttl), Version: version}
	return nil
}
func (c *Cache) Get(key string) (any, int, bool) {
	c.mu.RLock()
	entry, ok := c.items[key]
	now := c.clock()
	c.mu.RUnlock()
	if !ok || !now.Before(entry.ExpiresAt) {
		if ok {
			c.Delete(key)
		}
		return nil, 0, false
	}
	return entry.Value, entry.Version, true
}
func (c *Cache) CompareAndSet(key string, value any, ttl time.Duration, expected, newVersion int) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	old, ok := c.items[key]
	if !ok && expected != 0 {
		return false
	}
	if ok && old.Version != expected && !(!c.clock().Before(old.ExpiresAt)) {
		return false
	}
	c.items[key] = Entry{Value: value, ExpiresAt: c.clock().Add(ttl), Version: newVersion}
	return true
}
func (c *Cache) Delete(key string) { c.mu.Lock(); defer c.mu.Unlock(); delete(c.items, key) }
func (c *Cache) Clear()            { c.mu.Lock(); defer c.mu.Unlock(); c.items = map[string]Entry{} }
func (c *Cache) Size() int         { c.mu.RLock(); defer c.mu.RUnlock(); return len(c.items) }
func (c *Cache) Warm(ctx context.Context, keys []string, load func(context.Context, string) (any, int, error), ttl time.Duration) error {
	for _, key := range keys {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		value, version, e := load(ctx, key)
		if e != nil {
			return e
		}
		if e = c.Set(key, value, ttl, version); e != nil {
			return e
		}
	}
	return nil
}
