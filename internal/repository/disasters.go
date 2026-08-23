package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/11DingKing/autumn-grain-resilience/internal/domain"
)

type DisasterRepo struct{ DB DB }

func (r DisasterRepo) CreateReport(ctx context.Context, v domain.DisasterReport) error {
	_, e := r.DB.ExecContext(ctx, "INSERT INTO disaster_reports(id,region_id,reporter_id,kind,status,description,created_at) VALUES(?,?,?,?,?,?,?)", v.ID, v.RegionID, v.ReporterID, v.Kind, v.Status, v.Description, fmtTime(v.CreatedAt))
	return e
}
func (r DisasterRepo) CreateCase(ctx context.Context, v domain.RecoveryCase) error {
	_, e := r.DB.ExecContext(ctx, "INSERT INTO recovery_cases(id,report_id,region_id,status,owner_id,estimated_loss,version) VALUES(?,?,?,?,?,?,?)", v.ID, v.ReportID, v.RegionID, v.Status, v.OwnerID, v.EstimatedLoss, v.Version)
	return e
}
func (r DisasterRepo) Case(ctx context.Context, id string) (domain.RecoveryCase, error) {
	var v domain.RecoveryCase
	var o sql.NullString
	e := r.DB.QueryRowContext(ctx, "SELECT id,report_id,region_id,status,owner_id,estimated_loss,version FROM recovery_cases WHERE id=?", id).Scan(&v.ID, &v.ReportID, &v.RegionID, &v.Status, &o, &v.EstimatedLoss, &v.Version)
	if errors.Is(e, sql.ErrNoRows) {
		return v, domain.ErrNotFound
	}
	v.OwnerID = o.String
	return v, e
}
func (r DisasterRepo) TransitionCase(ctx context.Context, id, from, to string, version int) (bool, error) {
	res, e := r.DB.ExecContext(ctx, "UPDATE recovery_cases SET status=?,version=version+1 WHERE id=? AND status=? AND version=?", to, id, from, version)
	if e != nil {
		return false, e
	}
	n, _ := res.RowsAffected()
	return n == 1, nil
}
