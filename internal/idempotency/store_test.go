package idempotency

import (
	"context"
	"github.com/11DingKing/autumn-grain-resilience/internal/storage"
	"testing"
	"time"
)

func TestStoreLifecycle(t *testing.T) {
	db, e := storage.Open(":memory:")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	if e = storage.Migrate(context.Background(), db); e != nil {
		t.Fatal(e)
	}
	s := Store{DB: db}
	h := Hash([]byte("request"))
	if e = s.Save(context.Background(), "k", h, "response"); e != nil {
		t.Fatal(e)
	}
	got, ok, e := s.Lookup(context.Background(), "k", h)
	if e != nil || !ok || got != "response" {
		t.Fatalf("%q %v %v", got, ok, e)
	}
	if _, ok, e = s.Lookup(context.Background(), "missing", h); e != nil || ok {
		t.Fatal("missing key")
	}
	if _, e = s.Prune(context.Background(), time.Now().Add(time.Hour)); e != nil {
		t.Fatal(e)
	}
}
