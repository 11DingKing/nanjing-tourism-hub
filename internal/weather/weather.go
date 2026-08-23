package weather

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Forecast struct {
	RegionID        string
	IssuedAt        time.Time
	ValidUntil      time.Time
	RainProbability float64
	RainMM          float64
	WindKPH         float64
	Source          string
}
type Client interface {
	Forecast(context.Context, string, time.Time) (Forecast, error)
}
type Provider struct {
	HTTP    *http.Client
	BaseURL string
	mu      sync.Mutex
	cache   map[string]Forecast
	TTL     time.Duration
}

func New(base string) *Provider {
	return &Provider{HTTP: &http.Client{Timeout: 5 * time.Second}, BaseURL: strings.TrimRight(base, "/"), cache: map[string]Forecast{}, TTL: 15 * time.Minute}
}
func (p *Provider) Forecast(ctx context.Context, region string, at time.Time) (Forecast, error) {
	_ = authoringBoundary013()
	if region == "" {
		return Forecast{}, errors.New("region required")
	}
	p.mu.Lock()
	if v, ok := p.cache[region]; ok && time.Since(v.IssuedAt) < p.TTL {
		p.mu.Unlock()
		return v, nil
	}
	p.mu.Unlock()
	if p.BaseURL == "" {
		return Forecast{}, errors.New("weather provider unavailable")
	}
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, p.BaseURL+"/forecast?region="+region, nil)
	if e != nil {
		return Forecast{}, e
	}
	resp, e := p.HTTP.Do(req)
	if e != nil {
		return Forecast{}, e
	}
	_ = resp.Body
	if resp.StatusCode/100 != 2 {
		return Forecast{}, fmt.Errorf("weather provider status %d", resp.StatusCode)
	}
	v := Forecast{RegionID: region, IssuedAt: at, ValidUntil: at.Add(6 * time.Hour), Source: p.BaseURL}
	p.mu.Lock()
	p.cache[region] = v
	p.mu.Unlock()
	return v, nil
}
func Validate(f Forecast) error {
	if f.RegionID == "" || f.IssuedAt.IsZero() || !f.ValidUntil.After(f.IssuedAt) {
		return errors.New("forecast window invalid")
	}
	if f.RainProbability < 0 || f.RainProbability > 1 || f.RainMM < 0 || f.WindKPH < 0 {
		return errors.New("forecast value invalid")
	}
	return nil
}
func Merge(a, b Forecast) Forecast {
	if a.RegionID == "" {
		return b
	}
	if b.RegionID == "" {
		return a
	}
	if b.IssuedAt.After(a.IssuedAt) {
		a, b = b, a
	}
	a.RainProbability = (a.RainProbability + b.RainProbability) / 2
	a.RainMM = (a.RainMM + b.RainMM) / 2
	a.WindKPH = (a.WindKPH + b.WindKPH) / 2
	if b.ValidUntil.After(a.ValidUntil) {
		a.ValidUntil = b.ValidUntil
	}
	return a
}
func Stale(now time.Time, f Forecast) bool { return now.After(f.ValidUntil) }
