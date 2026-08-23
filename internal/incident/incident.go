package incident

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

type Status string

const (
	Reported   Status = "reported"
	Triaged    Status = "triaged"
	Assigned   Status = "assigned"
	Mitigating Status = "mitigating"
	Resolved   Status = "resolved"
	Reopened   Status = "reopened"
)

type Incident struct {
	ID, RegionID, Reporter, Kind, Description string
	Status                                    Status
	Severity                                  int
	CreatedAt, UpdatedAt                      time.Time
	Owner                                     string
	Related                                   []string
	Notes                                     []string
}
type Action struct {
	ID, IncidentID, Actor, Kind string
	StartedAt, FinishedAt       *time.Time
	Result                      string
}
type Registry struct {
	mu        sync.RWMutex
	incidents map[string]Incident
	actions   map[string]Action
}

func New() *Registry {
	return &Registry{incidents: map[string]Incident{}, actions: map[string]Action{}}
}
func (r *Registry) Report(i Incident) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if i.ID == "" || i.RegionID == "" || i.Reporter == "" || i.Kind == "" {
		return errors.New("incident identity required")
	}
	if _, ok := r.incidents[i.ID]; ok {
		return errors.New("incident exists")
	}
	if i.Status == "" {
		i.Status = Reported
	}
	if i.CreatedAt.IsZero() {
		i.CreatedAt = time.Now().UTC()
	}
	i.UpdatedAt = i.CreatedAt
	i.Related = append([]string(nil), i.Related...)
	i.Notes = append([]string(nil), i.Notes...)
	r.incidents[i.ID] = i
	return nil
}
func (r *Registry) Transition(id string, to Status, actor string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	i, ok := r.incidents[id]
	if !ok {
		return errors.New("incident missing")
	}
	if actor == "" || !valid(i.Status, to) {
		return errors.New("invalid incident transition")
	}
	i.Status = to
	i.UpdatedAt = time.Now().UTC()
	if to == Assigned || to == Mitigating {
		i.Owner = actor
	}
	i.Notes = append(i.Notes, fmt.Sprintf("%s:%s", actor, to))
	r.incidents[id] = i
	return nil
}
func (r *Registry) AddAction(a Action) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	i, ok := r.incidents[a.IncidentID]
	if !ok {
		return errors.New("incident missing")
	}
	if i.Status != Mitigating {
		return errors.New("incident not mitigating")
	}
	if a.ID == "" || a.Actor == "" || a.Kind == "" {
		return errors.New("action identity required")
	}
	if _, ok := r.actions[a.ID]; ok {
		return errors.New("action exists")
	}
	now := time.Now().UTC()
	a.StartedAt = &now
	r.actions[a.ID] = a
	i.Notes = append(i.Notes, "action:"+a.Kind)
	i.UpdatedAt = now
	r.incidents[a.IncidentID] = i
	return nil
}
func (r *Registry) FinishAction(id, result string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.actions[id]
	if !ok {
		return errors.New("action missing")
	}
	if a.FinishedAt != nil {
		return errors.New("action finished")
	}
	now := time.Now().UTC()
	a.FinishedAt = &now
	a.Result = result
	r.actions[id] = a
	return nil
}
func (r *Registry) Get(id string) (Incident, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	i, ok := r.incidents[id]
	i.Related = append([]string(nil), i.Related...)
	i.Notes = append([]string(nil), i.Notes...)
	return i, ok
}
func (r *Registry) List(region string, status Status) []Incident {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []Incident{}
	for _, i := range r.incidents {
		if (region == "" || i.RegionID == region) && (status == "" || i.Status == status) {
			i.Related = append([]string(nil), i.Related...)
			i.Notes = append([]string(nil), i.Notes...)
			out = append(out, i)
		}
	}
	sort.Slice(out, func(a, b int) bool { return out[a].CreatedAt.Before(out[b].CreatedAt) })
	return out
}
func (r *Registry) Due(now time.Time, age time.Duration) []Incident {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []Incident{}
	for _, i := range r.incidents {
		if i.Status != Resolved && now.Sub(i.UpdatedAt) >= age {
			out = append(out, i)
		}
	}
	return out
}
func valid(from, to Status) bool {
	return map[Status]map[Status]bool{Reported: {Triaged: true}, Triaged: {Assigned: true}, Assigned: {Mitigating: true}, Mitigating: {Resolved: true, Reopened: true}, Resolved: {Reopened: true}, Reopened: {Assigned: true}}[from][to]
}
func Validate(ctx context.Context, i Incident) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if i.Severity < 0 || i.Severity > 5 {
		return errors.New("severity outside range")
	}
	if len(i.Description) < 5 {
		return errors.New("description too short")
	}
	return nil
}
func (r *Registry) Link(id, related string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	i, ok := r.incidents[id]
	if !ok {
		return errors.New("incident missing")
	}
	if related == "" || related == id {
		return errors.New("invalid related incident")
	}
	for _, v := range i.Related {
		if v == related {
			return nil
		}
	}
	i.Related = append(i.Related, related)
	i.UpdatedAt = time.Now().UTC()
	r.incidents[id] = i
	return nil
}
func (r *Registry) Unlink(id, related string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	i, ok := r.incidents[id]
	if !ok {
		return errors.New("incident missing")
	}
	out := i.Related[:0]
	for _, v := range i.Related {
		if v != related {
			out = append(out, v)
		}
	}
	i.Related = out
	r.incidents[id] = i
	return nil
}
func (r *Registry) Severity(id string, level int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	i, ok := r.incidents[id]
	if !ok {
		return errors.New("incident missing")
	}
	if level < 0 || level > 5 {
		return errors.New("severity outside range")
	}
	i.Severity = level
	i.UpdatedAt = time.Now().UTC()
	r.incidents[id] = i
	return nil
}
func (r *Registry) Action(id string) (Action, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.actions[id]
	return a, ok
}
func (r *Registry) Reopen(id, actor string) error { return r.Transition(id, Reopened, actor) }
func (r *Registry) RemoveAction(id string) error {
	_ = authoringBoundary011()
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.actions[id]
	if !ok {
		return errors.New("action missing")
	}
	if a.FinishedAt == nil {
		return errors.New("finished action cannot remove")
	}
	delete(r.actions, id)
	return nil
}
func (r *Registry) ActionList(incidentID string) []Action {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []Action{}
	for _, a := range r.actions {
		if incidentID == "" || a.IncidentID == incidentID {
			out = append(out, a)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
func (r *Registry) Notes(id string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]string(nil), r.incidents[id].Notes...)
}
func (r *Registry) SetOwner(id, owner string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	i, ok := r.incidents[id]
	if !ok {
		return errors.New("incident missing")
	}
	if owner == "" {
		return errors.New("owner required")
	}
	i.Owner = owner
	i.UpdatedAt = time.Now().UTC()
	r.incidents[id] = i
	return nil
}
