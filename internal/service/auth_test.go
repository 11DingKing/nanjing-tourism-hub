package service

import (
	"context"
	"github.com/11DingKing/autumn-grain-resilience/internal/config"
	"github.com/11DingKing/autumn-grain-resilience/internal/storage"
	"testing"
	"time"
)

func TestAuthenticationLifecycle(t *testing.T) {
	db, err := storage.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = storage.Migrate(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{SessionTTL: time.Hour}
	s := NewRegistry(db, nil, cfg).Auth
	u, err := s.Register(context.Background(), "Lin", "agronomist", "secret")
	if err != nil {
		t.Fatal(err)
	}
	sess, err := s.Login(context.Background(), u.ID, "secret")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Authenticate(context.Background(), sess.ID); err != nil {
		t.Fatal(err)
	}
	if err = s.Logout(context.Background(), sess.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Authenticate(context.Background(), sess.ID); err == nil {
		t.Fatal("revoked session authenticated")
	}
}
