package reconciliation

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

type Entry struct {
	ID, RegionID, Reference string
	Expected, Actual        int64
	RecordedAt              time.Time
	State                   string
	Notes                   []string
}
type Result struct {
	EntryID string
	Delta   int64
	Matched bool
	Reason  string
}
type Book struct {
	mu      sync.Mutex
	entries map[string]Entry
}

func New() *Book { return &Book{entries: map[string]Entry{}} }
func (b *Book) Record(e Entry) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if e.ID == "" || e.RegionID == "" || e.Reference == "" {
		return errors.New("entry identity required")
	}
	if _, ok := b.entries[e.ID]; ok {
		return errors.New("entry exists")
	}
	if e.RecordedAt.IsZero() {
		e.RecordedAt = time.Now().UTC()
	}
	e.State = "open"
	e.Notes = append([]string(nil), e.Notes...)
	b.entries[e.ID] = e
	return nil
}
func (b *Book) Reconcile(id string, tolerance int64) Result {
	b.mu.Lock()
	defer b.mu.Unlock()
	e, ok := b.entries[id]
	if !ok {
		return Result{EntryID: id, Reason: "entry not found"}
	}
	d := e.Actual - e.Expected
	matched := d <= tolerance && d >= -tolerance
	if matched {
		e.State = "matched"
	} else {
		e.State = "exception"
		e.Notes = append(e.Notes, fmt.Sprintf("delta=%d", d))
	}
	b.entries[id] = e
	reason := "within tolerance"
	if !matched {
		reason = "outside tolerance"
	}
	return Result{EntryID: id, Delta: d, Matched: matched, Reason: reason}
}
func (b *Book) Batch(ids []string, tolerance int64) []Result {
	out := make([]Result, 0, len(ids))
	for _, id := range ids {
		out = append(out, b.Reconcile(id, tolerance))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].EntryID < out[j].EntryID })
	return out
}
func (b *Book) Exceptions(region string) []Entry {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := []Entry{}
	for _, e := range b.entries {
		if e.State == "exception" && (region == "" || e.RegionID == region) {
			e.Notes = append([]string(nil), e.Notes...)
			out = append(out, e)
		}
	}
	return out
}
func (b *Book) Balance(region string) int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	var total int64
	for _, e := range b.entries {
		if region == "" || e.RegionID == region {
			total += e.Actual - e.Expected
		}
	}
	return total
}
func Validate(e Entry) error {
	if e.Expected < 0 || e.Actual < 0 {
		return errors.New("amount cannot be negative")
	}
	if math.MaxInt64-e.Expected < e.Actual {
		return errors.New("amount overflow")
	}
	return nil
}
