package sqlite

import (
	"context"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
)

func (t *txStore) GetSupplyLot(ctx context.Context, id string) (domain.SupplyLot, error) {
	var v domain.SupplyLot
	var expires, updated string
	err := t.tx.QueryRowContext(ctx, `SELECT id,region_id,kind,quantity,reserved,issued,expires_at,status,version,updated_at FROM supply_lots WHERE id=?`, id).Scan(&v.ID, &v.RegionID, &v.Kind, &v.Quantity, &v.Reserved, &v.Issued, &expires, &v.Status, &v.Version, &updated)
	if err != nil {
		return domain.SupplyLot{}, translate("get", "supply_lot", id, err)
	}
	v.ExpiresAt = parseStamp(expires)
	v.UpdatedAt = parseStamp(updated)
	return v, nil
}
func (t *txStore) UpdateSupplyLot(ctx context.Context, v domain.SupplyLot, expected int64) error {
	result, err := t.tx.ExecContext(ctx, `UPDATE supply_lots SET reserved=?,issued=?,status=?,version=version+1,updated_at=? WHERE id=? AND version=? AND reserved+issued<=quantity`, v.Reserved, v.Issued, v.Status, stamp(v.UpdatedAt), v.ID, expected)
	if err != nil {
		return translate("update", "supply_lot", v.ID, err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return domain.VersionConflict("supply_lot", v.ID, expected, v.Version)
	}
	return nil
}
func (t *txStore) InsertAllocation(ctx context.Context, v domain.Allocation) error {
	_, err := t.tx.ExecContext(ctx, `INSERT INTO allocations(id,lot_id,incident_id,region_id,task_id,quantity,status,idempotency_key,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, v.ID, v.LotID, v.IncidentID, v.RegionID, v.TaskID, v.Quantity, v.Status, v.IdempotencyKey, v.Version, stamp(v.CreatedAt), stamp(v.UpdatedAt))
	return translate("insert", "allocation", v.ID, err)
}
func (t *txStore) GetAllocation(ctx context.Context, id string) (domain.Allocation, error) {
	var v domain.Allocation
	var created, updated string
	err := t.tx.QueryRowContext(ctx, `SELECT id,lot_id,incident_id,region_id,task_id,quantity,status,idempotency_key,version,created_at,updated_at FROM allocations WHERE id=?`, id).Scan(&v.ID, &v.LotID, &v.IncidentID, &v.RegionID, &v.TaskID, &v.Quantity, &v.Status, &v.IdempotencyKey, &v.Version, &created, &updated)
	if err != nil {
		return domain.Allocation{}, translate("get", "allocation", id, err)
	}
	v.CreatedAt = parseStamp(created)
	v.UpdatedAt = parseStamp(updated)
	return v, nil
}
func (t *txStore) UpdateAllocation(ctx context.Context, v domain.Allocation, expected int64) error {
	result, err := t.tx.ExecContext(ctx, `UPDATE allocations SET quantity=?,status=?,version=version+1,updated_at=? WHERE id=? AND version=?`, v.Quantity, v.Status, stamp(v.UpdatedAt), v.ID, expected)
	if err != nil {
		return translate("update", "allocation", v.ID, err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return domain.VersionConflict("allocation", v.ID, expected, v.Version)
	}
	return nil
}
