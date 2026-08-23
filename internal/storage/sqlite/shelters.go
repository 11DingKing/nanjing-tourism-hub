package sqlite

import (
	"context"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
)

func (t *txStore) GetShelter(ctx context.Context, id string) (domain.Shelter, error) {
	var v domain.Shelter
	var updated string
	err := t.tx.QueryRowContext(ctx, `SELECT id,region_id,name,capacity,reserved,occupied,status,version,updated_at FROM shelters WHERE id=?`, id).Scan(&v.ID, &v.RegionID, &v.Name, &v.Capacity, &v.Reserved, &v.Occupied, &v.Status, &v.Version, &updated)
	if err != nil {
		return domain.Shelter{}, translate("get", "shelter", id, err)
	}
	v.UpdatedAt = parseStamp(updated)
	return v, nil
}
func (t *txStore) UpdateShelter(ctx context.Context, v domain.Shelter, expected int64) error {
	result, err := t.tx.ExecContext(ctx, `UPDATE shelters SET reserved=?,occupied=?,status=?,version=version+1,updated_at=? WHERE id=? AND version=? AND reserved+occupied<=capacity`, v.Reserved, v.Occupied, v.Status, stamp(v.UpdatedAt), v.ID, expected)
	if err != nil {
		return translate("update", "shelter", v.ID, err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return domain.VersionConflict("shelter", v.ID, expected, v.Version)
	}
	return nil
}
func (t *txStore) InsertReservation(ctx context.Context, v domain.Reservation) error {
	_, err := t.tx.ExecContext(ctx, `INSERT INTO reservations(id,shelter_id,incident_id,region_id,people,status,idempotency_key,expires_at,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, v.ID, v.ShelterID, v.IncidentID, v.RegionID, v.People, v.Status, v.IdempotencyKey, stamp(v.ExpiresAt), v.Version, stamp(v.CreatedAt), stamp(v.UpdatedAt))
	return translate("insert", "reservation", v.ID, err)
}
func (t *txStore) GetReservation(ctx context.Context, id string) (domain.Reservation, error) {
	var v domain.Reservation
	var expires, created, updated string
	err := t.tx.QueryRowContext(ctx, `SELECT id,shelter_id,incident_id,region_id,people,status,idempotency_key,expires_at,version,created_at,updated_at FROM reservations WHERE id=?`, id).Scan(&v.ID, &v.ShelterID, &v.IncidentID, &v.RegionID, &v.People, &v.Status, &v.IdempotencyKey, &expires, &v.Version, &created, &updated)
	if err != nil {
		return domain.Reservation{}, translate("get", "reservation", id, err)
	}
	v.ExpiresAt = parseStamp(expires)
	v.CreatedAt = parseStamp(created)
	v.UpdatedAt = parseStamp(updated)
	return v, nil
}
func (t *txStore) UpdateReservation(ctx context.Context, v domain.Reservation, expected int64) error {
	result, err := t.tx.ExecContext(ctx, `UPDATE reservations SET people=?,status=?,expires_at=?,version=version+1,updated_at=? WHERE id=? AND version=?`, v.People, v.Status, stamp(v.ExpiresAt), stamp(v.UpdatedAt), v.ID, expected)
	if err != nil {
		return translate("update", "reservation", v.ID, err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return domain.VersionConflict("reservation", v.ID, expected, v.Version)
	}
	return nil
}
