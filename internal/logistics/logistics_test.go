package logistics

import (
	"context"
	"testing"
)

func TestInventoryMovements(t *testing.T) {
	s := New()
	if e := s.Add(Item{ID: "seed", RegionID: "r", Name: "seed", Unit: "kg", Quantity: 100, ReorderPoint: 20}); e != nil {
		t.Fatal(e)
	}
	if e := s.Move(Movement{ItemID: "seed", Kind: "reserve", Quantity: 30}); e != nil {
		t.Fatal(e)
	}
	if e := s.Move(Movement{ItemID: "seed", Kind: "consume", Quantity: 80}); e == nil {
		t.Fatal("reserved stock consumed")
	}
	if e := s.Move(Movement{ItemID: "seed", Kind: "release", Quantity: 30}); e != nil {
		t.Fatal(e)
	}
	if e := s.Move(Movement{ItemID: "seed", Kind: "consume", Quantity: 80}); e != nil {
		t.Fatal(e)
	}
	if len(s.LowStock("r")) != 1 {
		t.Fatal("low stock")
	}
}
func TestBatchContext(t *testing.T) {
	s := New()
	_ = s.Add(Item{ID: "a", RegionID: "r", Name: "a", Unit: "u", Quantity: 2})
	_ = s.Add(Item{ID: "b", RegionID: "r", Name: "b", Unit: "u", Quantity: 2})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if e := s.ReserveBatch(ctx, map[string]float64{"a": 1}); e == nil {
		t.Fatal("cancel ignored")
	}
}
