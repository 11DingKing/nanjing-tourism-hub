package metrics

import (
	"sync"
	"testing"
	"time"
)

func TestMetricsConcurrency(t *testing.T) {
	c := NewCounter()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				c.Add("requests", 1)
			}
		}()
	}
	wg.Wait()
	if c.Get("requests") != 5000 {
		t.Fatalf("count=%d", c.Get("requests"))
	}
	s := c.Snapshot()
	s["requests"] = 0
	if c.Get("requests") != 5000 {
		t.Fatal("snapshot aliased")
	}
	h := NewHistogram()
	h.Observe("latency", time.Second)
	h.Observe("latency", 2*time.Second)
	if h.Quantile("latency", 1) != 2*time.Second {
		t.Fatal("quantile")
	}
}
