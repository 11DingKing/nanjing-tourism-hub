package cache

import (
	"context"
	"testing"
	"time"
)

func TestCacheVersionsAndExpiry(t *testing.T) {
	now := time.Now()
	c := New()
	c.clock = func() time.Time { return now }
	if e := c.Set("k", "v", time.Hour, 1); e != nil {
		t.Fatal(e)
	}
	if v, ver, ok := c.Get("k"); !ok || v != "v" || ver != 1 {
		t.Fatal("get")
	}
	if !c.CompareAndSet("k", "v2", time.Hour, 1, 2) {
		t.Fatal("cas")
	}
	if c.CompareAndSet("k", "v3", time.Hour, 1, 3) {
		t.Fatal("stale cas")
	}
	now = now.Add(2 * time.Hour)
	if _, _, ok := c.Get("k"); ok {
		t.Fatal("expired")
	}
}
func TestCacheWarmContext(t *testing.T) {
	c := New()
	e := c.Warm(context.Background(), []string{"a", "b"}, func(_ context.Context, k string) (any, int, error) { return k, 1, nil }, time.Minute)
	if e != nil || c.Size() != 2 {
		t.Fatal(e)
	}
}
