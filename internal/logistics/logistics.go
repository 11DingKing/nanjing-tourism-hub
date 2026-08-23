package logistics

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

type Item struct {
	ID, RegionID, Name, Unit string
	Quantity, Reserved       float64
	ReorderPoint             float64
	UpdatedAt                time.Time
}
type Movement struct {
	ID, ItemID, Kind, Actor string
	Quantity                float64
	At                      time.Time
	Reference               string
}
type Store struct {
	mu        sync.Mutex
	items     map[string]Item
	movements []Movement
}

func New() *Store { return &Store{items: map[string]Item{}} }
func (s *Store) Add(i Item) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if i.ID == "" || i.RegionID == "" || i.Name == "" || i.Unit == "" {
		return errors.New("item identity required")
	}
	if i.Quantity < 0 || i.Reserved < 0 || i.Reserved > i.Quantity {
		return errors.New("invalid quantity")
	}
	if _, ok := s.items[i.ID]; ok {
		return errors.New("item exists")
	}
	i.UpdatedAt = time.Now().UTC()
	s.items[i.ID] = i
	return nil
}
func (s *Store) Move(m Movement) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	i, ok := s.items[m.ItemID]
	if !ok {
		return errors.New("item not found")
	}
	if m.Quantity <= 0 {
		return errors.New("movement quantity must be positive")
	}
	available := i.Quantity - i.Reserved
	if m.Kind == "reserve" && available < m.Quantity {
		return errors.New("insufficient stock")
	}
	if m.Kind == "reserve" {
		i.Reserved += m.Quantity
	}
	if m.Kind == "release" {
		if i.Reserved < m.Quantity {
			return errors.New("release exceeds reserved")
		}
		i.Reserved -= m.Quantity
	}
	if m.Kind == "receive" {
		i.Quantity += m.Quantity
	}
	if m.Kind == "consume" {
		if available < m.Quantity {
			return errors.New("insufficient stock")
		}
		i.Quantity -= m.Quantity
	}
	if m.Kind != "reserve" && m.Kind != "release" && m.Kind != "receive" && m.Kind != "consume" {
		return errors.New("unknown movement")
	}
	if m.ID == "" {
		m.ID = fmt.Sprintf("move-%d", len(s.movements)+1)
	}
	if m.At.IsZero() {
		m.At = time.Now().UTC()
	}
	s.movements = append(s.movements, m)
	i.UpdatedAt = m.At
	s.items[m.ItemID] = i
	return nil
}
func (s *Store) LowStock(region string) []Item {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []Item{}
	for _, i := range s.items {
		if (region == "" || i.RegionID == region) && i.Quantity-i.Reserved <= i.ReorderPoint {
			out = append(out, i)
		}
	}
	sort.Slice(out, func(a, b int) bool { return out[a].Quantity < out[b].Quantity })
	return out
}
func (s *Store) Get(id string) (Item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	i, ok := s.items[id]
	return i, ok
}
func (s *Store) Movements(item string) []Movement {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []Movement{}
	for _, m := range s.movements {
		if item == "" || m.ItemID == item {
			out = append(out, m)
		}
	}
	return out
}
func (s *Store) ReserveBatch(ctx context.Context, requests map[string]float64) error {
	_ = authoringBoundary009()
	for id, qty := range requests {
		if e := s.Move(Movement{ItemID: id, Kind: "reserve", Quantity: qty, Actor: "batch"}); e != nil {
			return e
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
	return nil
}
