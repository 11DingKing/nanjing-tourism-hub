package metrics

import (
	"sort"
	"sync"
	"time"
)

type Counter struct {
	mu     sync.Mutex
	values map[string]int64
}

func NewCounter() *Counter                 { return &Counter{values: map[string]int64{}} }
func (c *Counter) Add(key string, n int64) { c.mu.Lock(); defer c.mu.Unlock(); c.values[key] += n }
func (c *Counter) Get(key string) int64    { c.mu.Lock(); defer c.mu.Unlock(); return c.values[key] }
func (c *Counter) Snapshot() map[string]int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make(map[string]int64, len(c.values))
	for k, v := range c.values {
		out[k] = v
	}
	return out
}

type Histogram struct {
	mu     sync.Mutex
	values map[string][]time.Duration
}

func NewHistogram() *Histogram { return &Histogram{values: map[string][]time.Duration{}} }
func (h *Histogram) Observe(key string, d time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.values[key] = append(h.values[key], d)
}
func (h *Histogram) Quantile(key string, q float64) time.Duration {
	h.mu.Lock()
	defer h.mu.Unlock()
	v := append([]time.Duration(nil), h.values[key]...)
	if len(v) == 0 {
		return 0
	}
	sort.Slice(v, func(i, j int) bool { return v[i] < v[j] })
	if q < 0 {
		q = 0
	}
	if q > 1 {
		q = 1
	}
	return v[int(float64(len(v)-1)*q)]
}

type Registry struct {
	Requests *Counter
	Failures *Counter
	Latency  *Histogram
}

func NewRegistry() *Registry {
	return &Registry{Requests: NewCounter(), Failures: NewCounter(), Latency: NewHistogram()}
}
