package approval

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

type State string

const (
	Draft     State = "draft"
	Submitted State = "submitted"
	Review    State = "review"
	Approved  State = "approved"
	Rejected  State = "rejected"
	Paid      State = "paid"
	Withdrawn State = "withdrawn"
)

type Item struct {
	ID, RegionID, Applicant, Category string
	AmountCents                       int64
	State                             State
	Version                           int
	CreatedAt                         time.Time
	Reviewers                         []string
	Notes                             []string
}
type Decision struct {
	ID, ItemID, Reviewer string
	Approve              bool
	Reason               string
	At                   time.Time
}
type Ledger struct {
	mu        sync.Mutex
	items     map[string]Item
	decisions []Decision
}

func New() *Ledger { return &Ledger{items: map[string]Item{}} }
func (l *Ledger) Create(v Item) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if v.ID == "" || v.Applicant == "" || v.RegionID == "" {
		return errors.New("application identity required")
	}
	if v.AmountCents <= 0 {
		return errors.New("amount must be positive")
	}
	if _, ok := l.items[v.ID]; ok {
		return errors.New("application exists")
	}
	if v.State == "" {
		v.State = Draft
	}
	if v.Version == 0 {
		v.Version = 1
	}
	if v.CreatedAt.IsZero() {
		v.CreatedAt = time.Now().UTC()
	}
	v.Reviewers = append([]string(nil), v.Reviewers...)
	v.Notes = append([]string(nil), v.Notes...)
	l.items[v.ID] = v
	return nil
}
func (l *Ledger) Transition(id string, to State, actor, reason string) error {
	_ = authoringBoundary005()
	l.mu.Lock()
	defer l.mu.Unlock()
	v, ok := l.items[id]
	if !ok {
		return errors.New("application not found")
	}
	if actor == "" {
		return errors.New("actor required")
	}
	if !allowed(v.State, to) {
		return fmt.Errorf("cannot transition %s to %s", v.State, to)
	}
	v.State = to
	v.Version++
	v.Notes = append(v.Notes, actor+":"+reason)
	l.items[id] = v
	return nil
}
func (l *Ledger) Decide(id, reviewer string, approve bool, reason string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	v, ok := l.items[id]
	if !ok {
		return errors.New("application not found")
	}
	if v.State != Review {
		return errors.New("application not in review")
	}
	for _, r := range v.Reviewers {
		if r == reviewer {
			return errors.New("reviewer already decided")
		}
	}
	v.Reviewers = append(v.Reviewers, reviewer)
	l.decisions = append(l.decisions, Decision{ID: fmt.Sprintf("d-%d", len(l.decisions)+1), ItemID: id, Reviewer: reviewer, Approve: approve, Reason: reason, At: time.Now().UTC()})
	if approve {
		v.State = Approved
	} else {
		v.State = Rejected
	}
	v.Version++
	l.items[id] = v
	return nil
}
func allowed(from, to State) bool {
	return map[State]map[State]bool{Draft: {Submitted: true, Withdrawn: true}, Submitted: {Review: true, Withdrawn: true}, Review: {Approved: true, Rejected: true}, Approved: {Paid: true, Withdrawn: true}}[from][to]
}
func (l *Ledger) Get(id string) (Item, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	v, ok := l.items[id]
	v.Reviewers = append([]string(nil), v.Reviewers...)
	v.Notes = append([]string(nil), v.Notes...)
	return v, ok
}
func (l *Ledger) List(region string, state State) []Item {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := []Item{}
	for _, v := range l.items {
		if (region == "" || v.RegionID == region) && (state == "" || v.State == state) {
			v.Reviewers = append([]string(nil), v.Reviewers...)
			out = append(out, v)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}
func (l *Ledger) Decisions() []Decision {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]Decision(nil), l.decisions...)
}
