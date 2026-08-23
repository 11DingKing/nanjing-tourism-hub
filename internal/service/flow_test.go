package service

import (
	"context"
	"github.com/11DingKing/autumn-grain-resilience/internal/config"
	"github.com/11DingKing/autumn-grain-resilience/internal/domain"
	"github.com/11DingKing/autumn-grain-resilience/internal/storage"
	"testing"
	"time"
)

func TestMonitoringCreatesAlertAndOutboxForHighSeverity(t *testing.T) {
	db, err := storage.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = storage.Migrate(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	s := NewRegistry(db, nil, config.Config{})
	u := domain.User{ID: "u1", Name: "observer", Role: "agronomist", PasswordHash: "x", CreatedAt: time.Now()}
	if err = s.Auth.Users.Create(context.Background(), u); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("INSERT INTO regions(id,name,flood_risk) VALUES('r1','North','high')"); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("INSERT INTO plots(id,region_id,owner_id,acres,crop,status,version) VALUES('p1','r1','f1',10,'MAIZE','active',1)"); err != nil {
		t.Fatal(err)
	}
	alert, err := s.Monitoring.Record(context.Background(), domain.Observation{ID: "o1", PlotID: "p1", ObserverID: "u1", Severity: "high", Kind: "flood", Notes: "waterlogged"}, "r1", "req")
	if err != nil {
		t.Fatal(err)
	}
	if alert.ID == "" {
		t.Fatal("missing alert")
	}
	var n int
	if err = db.QueryRow("SELECT COUNT(*) FROM outbox_events WHERE status='pending'").Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("outbox=%d", n)
	}
}
