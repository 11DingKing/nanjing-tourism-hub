package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/11DingKing/autumn-grain-resilience/internal/domain"
)

type RegionRepo struct{ DB DB }

func (r RegionRepo) Create(ctx context.Context, v domain.Region) error {
	_, e := r.DB.ExecContext(ctx, "INSERT INTO regions(id,name,parent_id,flood_risk) VALUES(?,?,?,?)", v.ID, v.Name, v.ParentID, v.FloodRisk)
	return e
}
func (r RegionRepo) ByID(ctx context.Context, id string) (domain.Region, error) {
	var v domain.Region
	var p sql.NullString
	e := r.DB.QueryRowContext(ctx, "SELECT id,name,parent_id,flood_risk FROM regions WHERE id=?", id).Scan(&v.ID, &v.Name, &p, &v.FloodRisk)
	if errors.Is(e, sql.ErrNoRows) {
		return v, domain.ErrNotFound
	}
	v.ParentID = p.String
	return v, e
}
func (r RegionRepo) Children(ctx context.Context, parent string) ([]domain.Region, error) {
	rows, e := r.DB.QueryContext(ctx, "SELECT id,name,parent_id,flood_risk FROM regions WHERE parent_id=? ORDER BY name", parent)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.Region{}
	for rows.Next() {
		var v domain.Region
		var p sql.NullString
		if e = rows.Scan(&v.ID, &v.Name, &p, &v.FloodRisk); e != nil {
			return nil, e
		}
		v.ParentID = p.String
		out = append(out, v)
	}
	return out, rows.Err()
}
