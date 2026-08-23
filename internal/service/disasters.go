package service

import (
	"context"
	"database/sql"
	"github.com/11DingKing/autumn-grain-resilience/internal/domain"
	"github.com/11DingKing/autumn-grain-resilience/internal/repository"
	"time"
)

type DisasterService struct {
	DB    *sql.DB
	Repo  repository.DisasterRepo
	Audit *AuditService
}

func (s *DisasterService) Report(ctx context.Context, v domain.DisasterReport, request string) (domain.RecoveryCase, error) {
	if v.Status == "" {
		v.Status = "reported"
	}
	if v.CreatedAt.IsZero() {
		v.CreatedAt = time.Now().UTC()
	}
	tx, e := s.DB.BeginTx(ctx, nil)
	if e != nil {
		return domain.RecoveryCase{}, e
	}
	defer tx.Rollback()
	if _, e = tx.ExecContext(ctx, "INSERT INTO disaster_reports(id,region_id,reporter_id,kind,status,description,created_at) VALUES(?,?,?,?,?,?,?)", v.ID, v.RegionID, v.ReporterID, v.Kind, v.Status, v.Description, v.CreatedAt.UTC().Format(time.RFC3339Nano)); e != nil {
		return domain.RecoveryCase{}, e
	}
	c := domain.RecoveryCase{ID: id("rec_"), ReportID: v.ID, RegionID: v.RegionID, Status: "triaged", EstimatedLoss: 0, Version: 1}
	if _, e = tx.ExecContext(ctx, "INSERT INTO recovery_cases(id,report_id,region_id,status,owner_id,estimated_loss,version) VALUES(?,?,?,?,?,?,?)", c.ID, c.ReportID, c.RegionID, c.Status, nil, c.EstimatedLoss, c.Version); e != nil {
		return domain.RecoveryCase{}, e
	}
	if e = tx.Commit(); e != nil {
		return domain.RecoveryCase{}, e
	}
	if e = s.Audit.Record(ctx, v.ReporterID, "report_disaster", "disaster_report", v.ID, "success", request); e != nil {
		return domain.RecoveryCase{}, e
	}
	return c, nil
}
func (s *DisasterService) Transition(ctx context.Context, id, from, to string, version int, actor, request string) error {
	if !domain.ValidReportTransition(from, to) && from != to {
		return domain.ErrInvalidState
	}
	ok, e := s.Repo.TransitionCase(ctx, id, from, to, version)
	if e != nil {
		return e
	}
	if !ok {
		return domain.ErrConflict
	}
	return s.Audit.Record(ctx, actor, "transition_recovery_case", "recovery_case", id, to, request)
}
