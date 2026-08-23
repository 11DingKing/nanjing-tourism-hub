package repository

import (
	"context"
	"github.com/11DingKing/autumn-grain-resilience/internal/domain"
)

type AidRepo struct{ DB DB }

func (r AidRepo) Create(ctx context.Context, v domain.AidCase) error {
	_, e := r.DB.ExecContext(ctx, "INSERT INTO aid_cases(id,region_id,farmer_id,status,reason,amount_cents,version) VALUES(?,?,?,?,?,?,?)", v.ID, v.RegionID, v.FarmerID, v.Status, v.Reason, v.AmountCents, v.Version)
	return e
}
func (r AidRepo) Transition(ctx context.Context, id, from, to string, version int) (bool, error) {
	res, e := r.DB.ExecContext(ctx, "UPDATE aid_cases SET status=?,version=version+1 WHERE id=? AND status=? AND version=?", to, id, from, version)
	if e != nil {
		return false, e
	}
	n, _ := res.RowsAffected()
	return n == 1, nil
}
