package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/11DingKing/autumn-grain-resilience/internal/domain"
)

type PlotRepo struct{ DB DB }

func (r PlotRepo) Create(ctx context.Context, v domain.Plot) error {
	_, e := r.DB.ExecContext(ctx, "INSERT INTO plots(id,region_id,owner_id,acres,crop,status,version) VALUES(?,?,?,?,?,?,?)", v.ID, v.RegionID, v.OwnerID, v.Acres, v.Crop, v.Status, v.Version)
	return e
}
func (r PlotRepo) ByID(ctx context.Context, id string) (domain.Plot, error) {
	var v domain.Plot
	e := r.DB.QueryRowContext(ctx, "SELECT id,region_id,owner_id,acres,crop,status,version FROM plots WHERE id=?", id).Scan(&v.ID, &v.RegionID, &v.OwnerID, &v.Acres, &v.Crop, &v.Status, &v.Version)
	if errors.Is(e, sql.ErrNoRows) {
		return v, domain.ErrNotFound
	}
	return v, e
}
func (r PlotRepo) UpdateStatus(ctx context.Context, id, status string, version int) (bool, error) {
	res, e := r.DB.ExecContext(ctx, "UPDATE plots SET status=?,version=version+1 WHERE id=? AND version=?", status, id, version)
	if e != nil {
		return false, e
	}
	n, _ := res.RowsAffected()
	return n == 1, nil
}
