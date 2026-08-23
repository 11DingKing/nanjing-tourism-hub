package repository

import (
	"context"
	"github.com/11DingKing/autumn-grain-resilience/internal/domain"
)

type DryingRepo struct{ DB DB }

func (r DryingRepo) CreateSite(ctx context.Context, v domain.DryingSite) error {
	_, e := r.DB.ExecContext(ctx, "INSERT INTO drying_sites(id,region_id,name,capacity_tons,active,version) VALUES(?,?,?,?,?,?)", v.ID, v.RegionID, v.Name, v.CapacityTons, v.Active, v.Version)
	return e
}
func (r DryingRepo) Reserve(ctx context.Context, v domain.DryingReservation) error {
	_, e := r.DB.ExecContext(ctx, "INSERT INTO drying_reservations(id,site_id,plot_id,status,tons,starts_at,ends_at,version) VALUES(?,?,?,?,?,?,?,?)", v.ID, v.SiteID, v.PlotID, v.Status, v.Tons, fmtTime(v.StartsAt), fmtTime(v.EndsAt), v.Version)
	return e
}
func (r DryingRepo) Used(ctx context.Context, site string, start, end string) (float64, error) {
	var used float64
	e := r.DB.QueryRowContext(ctx, "SELECT COALESCE(SUM(tons),0) FROM drying_reservations WHERE site_id=? AND status IN ('held','confirmed') AND starts_at < ? AND ends_at > ?", site, end, start).Scan(&used)
	return used, e
}
