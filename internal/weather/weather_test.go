package weather

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProviderCacheAndContext(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; w.WriteHeader(200) }))
	defer server.Close()
	p := New(server.URL)
	now := time.Now()
	f, e := p.Forecast(context.Background(), "r1", now)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = p.Forecast(context.Background(), "r1", now); e != nil || calls != 1 {
		t.Fatalf("cache calls=%d", calls)
	}
	if e = Validate(f); e != nil {
		t.Fatal(e)
	}
	if !Stale(now.Add(7*time.Hour), f) {
		t.Fatal("stale")
	}
}
func TestProviderErrorsAndMerge(t *testing.T) {
	p := New("")
	if _, e := p.Forecast(context.Background(), "r", time.Now()); e == nil {
		t.Fatal("empty provider accepted")
	}
	a := Forecast{RegionID: "r", IssuedAt: time.Now(), ValidUntil: time.Now().Add(time.Hour), RainMM: 10}
	b := Forecast{RegionID: "r", IssuedAt: a.IssuedAt.Add(time.Minute), ValidUntil: a.ValidUntil.Add(time.Hour), RainMM: 20}
	if got := Merge(a, b); got.RainMM != 15 {
		t.Fatal(got)
	}
}
