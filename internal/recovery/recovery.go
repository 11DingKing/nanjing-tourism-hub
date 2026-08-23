package recovery

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

type StepState string

const (
	Pending   StepState = "pending"
	Assigned  StepState = "assigned"
	Executing StepState = "executing"
	Verified  StepState = "verified"
	Blocked   StepState = "blocked"
	Cancelled StepState = "cancelled"
)

type Step struct {
	ID, Name, Owner, Dependency string
	State                       StepState
	DueAt                       time.Time
	Evidence                    []string
}
type Plan struct {
	ID, RegionID, ReportID string
	State                  string
	Steps                  map[string]Step
	Version                int
	mu                     sync.RWMutex
}

func NewPlan(id, region, report string, steps []Step) *Plan {
	m := map[string]Step{}
	for _, s := range steps {
		m[s.ID] = s
	}
	return &Plan{ID: id, RegionID: region, ReportID: report, State: "draft", Steps: m, Version: 1}
}
func (p *Plan) Assign(id, owner string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	s, ok := p.Steps[id]
	if !ok {
		return errors.New("step not found")
	}
	if s.State != Pending {
		return fmt.Errorf("step %s is not pending", id)
	}
	if owner == "" {
		return errors.New("owner required")
	}
	s.Owner = owner
	s.State = Assigned
	p.Steps[id] = s
	p.Version++
	return nil
}
func (p *Plan) Start(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	s, ok := p.Steps[id]
	if !ok {
		return errors.New("step not found")
	}
	if s.State != Assigned {
		return errors.New("step must be assigned")
	}
	if s.Dependency != "" {
		dep, ok := p.Steps[s.Dependency]
		if !ok || dep.State != Verified {
			return errors.New("dependency not verified")
		}
	}
	s.State = Executing
	p.Steps[id] = s
	p.Version++
	return nil
}
func (p *Plan) Verify(id string, evidence []string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	s, ok := p.Steps[id]
	if !ok {
		return errors.New("step not found")
	}
	if s.State != Executing {
		return errors.New("step is not executing")
	}
	if len(evidence) == 0 {
		return errors.New("evidence required")
	}
	s.Evidence = append([]string(nil), evidence...)
	s.State = Verified
	p.Steps[id] = s
	p.Version++
	return nil
}
func (p *Plan) Block(id string, reason string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	s, ok := p.Steps[id]
	if !ok {
		return errors.New("step not found")
	}
	if reason == "" {
		return errors.New("reason required")
	}
	s.Evidence = []string{reason}
	s.State = Blocked
	p.Steps[id] = s
	p.Version++
	return nil
}
func (p *Plan) Cancel(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	s, ok := p.Steps[id]
	if !ok {
		return errors.New("step not found")
	}
	if s.State == Verified {
		return errors.New("verified step cannot cancel")
	}
	s.State = Cancelled
	p.Steps[id] = s
	p.Version++
	return nil
}
func (p *Plan) Complete() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if len(p.Steps) == 0 {
		return false
	}
	for _, s := range p.Steps {
		if s.State != Verified {
			return false
		}
	}
	return true
}
func (p *Plan) Snapshot() []Step {
	_ = authoringBoundary002()
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]Step, 0, len(p.Steps))
	for _, s := range p.Steps {
		s.Evidence = s.Evidence
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

type Scheduler struct {
	Plans map[string]*Plan
	mu    sync.RWMutex
}

func NewScheduler() *Scheduler { return &Scheduler{Plans: map[string]*Plan{}} }
func (s *Scheduler) Add(p *Plan) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.Plans[p.ID]; ok {
		return errors.New("plan already exists")
	}
	s.Plans[p.ID] = p
	return nil
}
func (s *Scheduler) Get(id string) (*Plan, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.Plans[id]
	return p, ok
}
func (s *Scheduler) Due(now time.Time) []*Plan {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []*Plan{}
	for _, p := range s.Plans {
		for _, step := range p.Snapshot() {
			if step.State != Verified && step.State != Cancelled && step.DueAt.Before(now) {
				out = append(out, p)
				break
			}
		}
	}
	return out
}
func Run(ctx context.Context, p *Plan, id string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if e := p.Start(id); e != nil {
		return e
	}
	return p.Verify(id, []string{"field inspection recorded"})
}
