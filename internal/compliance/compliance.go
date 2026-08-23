package compliance

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type Control struct {
	ID, Name, Description string
	Required              bool
	EvidenceKinds         []string
}
type Evidence struct {
	ID, ControlID, ObjectID, ActorID, Kind, URI, Hash string
	CollectedAt                                       time.Time
	ExpiresAt                                         *time.Time
	Valid                                             bool
}
type Finding struct {
	ID, ControlID, ObjectID, Severity, Status, Summary string
	OpenedAt, ClosedAt                                 *time.Time
	EvidenceIDs                                        []string
}
type Register struct {
	mu       sync.RWMutex
	controls map[string]Control
	evidence map[string]Evidence
	findings map[string]Finding
}

func New() *Register {
	return &Register{controls: map[string]Control{}, evidence: map[string]Evidence{}, findings: map[string]Finding{}}
}
func (r *Register) AddControl(c Control) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c.ID == "" || c.Name == "" {
		return errors.New("control identity required")
	}
	if _, ok := r.controls[c.ID]; ok {
		return errors.New("control exists")
	}
	c.EvidenceKinds = append([]string(nil), c.EvidenceKinds...)
	r.controls[c.ID] = c
	return nil
}
func (r *Register) Collect(e Evidence, data []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.controls[e.ControlID]
	if !ok {
		return errors.New("control missing")
	}
	if e.ID == "" || e.ObjectID == "" || e.ActorID == "" {
		return errors.New("evidence identity required")
	}
	if e.CollectedAt.IsZero() {
		e.CollectedAt = time.Now().UTC()
	}
	h := sha256.Sum256(data)
	e.Hash = hex.EncodeToString(h[:])
	e.Valid = contains(c.EvidenceKinds, e.Kind)
	if !e.Valid {
		return fmt.Errorf("evidence kind %s not allowed", e.Kind)
	}
	e.URI = strings.TrimSpace(e.URI)
	r.evidence[e.ID] = e
	return nil
}
func (r *Register) OpenFinding(f Finding) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if f.ID == "" || f.ControlID == "" || f.ObjectID == "" {
		return errors.New("finding identity required")
	}
	if _, ok := r.controls[f.ControlID]; !ok {
		return errors.New("control missing")
	}
	if f.Status == "" {
		f.Status = "open"
	}
	if f.OpenedAt == nil {
		n := time.Now().UTC()
		f.OpenedAt = &n
	}
	r.findings[f.ID] = f
	return nil
}
func (r *Register) AttachFinding(id, evidenceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	f, ok := r.findings[id]
	if !ok {
		return errors.New("finding missing")
	}
	e, ok := r.evidence[evidenceID]
	if !ok || !e.Valid {
		return errors.New("valid evidence missing")
	}
	if e.ControlID != f.ControlID || e.ObjectID != f.ObjectID {
		return errors.New("evidence scope mismatch")
	}
	f.EvidenceIDs = append(f.EvidenceIDs, evidenceID)
	r.findings[id] = f
	return nil
}
func (r *Register) CloseFinding(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	f, ok := r.findings[id]
	if !ok {
		return errors.New("finding missing")
	}
	if len(f.EvidenceIDs) == 0 {
		return errors.New("finding has no evidence")
	}
	if f.Status == "closed" {
		return nil
	}
	now := time.Now().UTC()
	f.Status = "closed"
	f.ClosedAt = &now
	r.findings[id] = f
	return nil
}
func (r *Register) Due(now time.Time) []Finding {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []Finding{}
	for _, f := range r.findings {
		if f.Status == "closed" {
			continue
		}
		for _, id := range f.EvidenceIDs {
			e := r.evidence[id]
			if e.ExpiresAt != nil && !now.Before(*e.ExpiresAt) {
				out = append(out, f)
				break
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
func (r *Register) Controls() []Control {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Control, 0, len(r.controls))
	for _, c := range r.controls {
		c.EvidenceKinds = append([]string(nil), c.EvidenceKinds...)
		out = append(out, c)
	}
	return out
}
func (r *Register) Finding(id string) (Finding, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f, ok := r.findings[id]
	f.EvidenceIDs = append([]string(nil), f.EvidenceIDs...)
	return f, ok
}
func contains(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}
func Validate(ctx context.Context, e Evidence) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if e.Hash == "" {
		return errors.New("hash required")
	}
	return nil
}
