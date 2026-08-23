package sqlite

import (
	"context"
	"database/sql"

	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
)

func (t *txStore) InsertIncident(ctx context.Context, v domain.Incident) error {
	_, err := t.tx.ExecContext(ctx, `INSERT INTO incidents(id,warning_id,name,status,commander_id,activated_at,closed_at,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?)`, v.ID, v.WarningID, v.Name, v.Status, v.CommanderID, stamp(v.ActivatedAt), nullableStamp(v.ClosedAt), v.Version, stamp(v.CreatedAt), stamp(v.UpdatedAt))
	return translate("insert", "incident", v.ID, err)
}
func (t *txStore) GetIncident(ctx context.Context, id string) (domain.Incident, error) {
	var v domain.Incident
	var activated, created, updated string
	var closed sql.NullString
	err := t.tx.QueryRowContext(ctx, `SELECT id,warning_id,name,status,commander_id,activated_at,closed_at,version,created_at,updated_at FROM incidents WHERE id=?`, id).Scan(&v.ID, &v.WarningID, &v.Name, &v.Status, &v.CommanderID, &activated, &closed, &v.Version, &created, &updated)
	if err != nil {
		return domain.Incident{}, translate("get", "incident", id, err)
	}
	v.ActivatedAt = parseStamp(activated)
	v.ClosedAt = scanOptional(closed)
	v.CreatedAt = parseStamp(created)
	v.UpdatedAt = parseStamp(updated)
	return v, nil
}
func (t *txStore) UpdateIncident(ctx context.Context, v domain.Incident, expected int64) error {
	result, err := t.tx.ExecContext(ctx, `UPDATE incidents SET status=?,commander_id=?,closed_at=?,version=version+1,updated_at=? WHERE id=? AND version=?`, v.Status, v.CommanderID, nullableStamp(v.ClosedAt), stamp(v.UpdatedAt), v.ID, expected)
	if err != nil {
		return translate("update", "incident", v.ID, err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return domain.VersionConflict("incident", v.ID, expected, v.Version)
	}
	return nil
}
