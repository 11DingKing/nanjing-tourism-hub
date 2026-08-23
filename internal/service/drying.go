package service

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/11DingKing/autumn-grain-resilience/internal/domain"
	"github.com/11DingKing/autumn-grain-resilience/internal/repository"
	"time"
)

type DryingService struct {
	DB    *sql.DB
	Repo  repository.DryingRepo
	Plots repository.PlotRepo
	Audit *AuditService
}

func (s *DryingService) Reserve(ctx context.Context, v domain.DryingReservation, capacity float64, request string) error {
	if v.StartsAt.Before(time.Now().Add(-time.Minute)) {
		return domain.ErrExpired
	}
	tx, e := s.DB.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var used float64
	if e = tx.QueryRowContext(ctx, "SELECT COALESCE(SUM(tons),0) FROM drying_reservations WHERE site_id=? AND status IN ('held','confirmed') AND starts_at < ? AND ends_at > ?", v.SiteID, v.EndsAt.UTC().Format(time.RFC3339Nano), v.StartsAt.UTC().Format(time.RFC3339Nano)).Scan(&used); e != nil {
		return e
	}
	if used+v.Tons > capacity {
		return fmt.Errorf("%w: %.2f requested with %.2f used", domain.ErrCapacity, v.Tons, used)
	}
	if v.Status == "" {
		v.Status = "held"
	}
	if v.Version == 0 {
		v.Version = 1
	}
	if _, e = tx.ExecContext(ctx, "INSERT INTO drying_reservations(id,site_id,plot_id,status,tons,starts_at,ends_at,version) VALUES(?,?,?,?,?,?,?,?)", v.ID, v.SiteID, v.PlotID, v.Status, v.Tons, v.StartsAt.UTC().Format(time.RFC3339Nano), v.EndsAt.UTC().Format(time.RFC3339Nano), v.Version); e != nil {
		return e
	}
	if e = tx.Commit(); e != nil {
		return e
	}
	return s.Audit.Record(ctx, "system", "reserve_drying", "drying_reservation", v.ID, "success", request)
}
