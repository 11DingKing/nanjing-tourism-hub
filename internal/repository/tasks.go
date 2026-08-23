package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/11DingKing/autumn-grain-resilience/internal/domain"
	"time"
)

type TaskRepo struct{ DB DB }

func (r TaskRepo) Create(ctx context.Context, v domain.FieldTask) error {
	_, e := r.DB.ExecContext(ctx, "INSERT INTO field_tasks(id,plot_id,assignee_id,kind,status,lease_id,due_at,version) VALUES(?,?,?,?,?,?,?,?)", v.ID, v.PlotID, v.AssigneeID, v.Kind, v.Status, v.LeaseID, fmtTime(v.DueAt), v.Version)
	return e
}
func (r TaskRepo) ByID(ctx context.Context, id string) (domain.FieldTask, error) {
	var v domain.FieldTask
	var a, l, d string
	e := r.DB.QueryRowContext(ctx, "SELECT id,plot_id,assignee_id,kind,status,lease_id,due_at,version FROM field_tasks WHERE id=?", id).Scan(&v.ID, &v.PlotID, &a, &v.Kind, &v.Status, &l, &d, &v.Version)
	if errors.Is(e, sql.ErrNoRows) {
		return v, domain.ErrNotFound
	}
	v.AssigneeID = a
	v.LeaseID = l
	v.DueAt = parseTime(d)
	return v, e
}
func (r TaskRepo) Claim(ctx context.Context, id, worker, lease string, until time.Time) (bool, error) {
	res, e := r.DB.ExecContext(ctx, "UPDATE field_tasks SET status='claimed',assignee_id=?,lease_id=?,version=version+1 WHERE id=? AND status='queued'", worker, lease, id)
	if e != nil {
		return false, e
	}
	n, _ := res.RowsAffected()
	return n == 1, nil
}
func (r TaskRepo) Transition(ctx context.Context, id, from, to string, version int) (bool, error) {
	res, e := r.DB.ExecContext(ctx, "UPDATE field_tasks SET status=?,version=version+1 WHERE id=? AND status=? AND version=?", to, id, from, version)
	if e != nil {
		return false, e
	}
	n, _ := res.RowsAffected()
	return n == 1, nil
}
