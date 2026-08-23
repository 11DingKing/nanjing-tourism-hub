package sqlite

import (
	"context"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
)

func (t *txStore) InsertResponse(ctx context.Context, v domain.Response) error {
	_, err := t.tx.ExecContext(ctx, `INSERT INTO responses(id,incident_id,region_id,level,status,activated_by,activated_at,reason,version,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?)`, v.ID, v.IncidentID, v.RegionID, v.Level, v.Status, v.ActivatedBy, stamp(v.ActivatedAt), v.Reason, v.Version, stamp(v.UpdatedAt))
	return translate("insert", "response", v.ID, err)
}
func (t *txStore) GetResponse(ctx context.Context, incidentID, regionID string) (domain.Response, error) {
	var v domain.Response
	var activated, updated string
	err := t.tx.QueryRowContext(ctx, `SELECT id,incident_id,region_id,level,status,activated_by,activated_at,reason,version,updated_at FROM responses WHERE incident_id=? AND region_id=?`, incidentID, regionID).Scan(&v.ID, &v.IncidentID, &v.RegionID, &v.Level, &v.Status, &v.ActivatedBy, &activated, &v.Reason, &v.Version, &updated)
	if err != nil {
		return domain.Response{}, translate("get", "response", incidentID+":"+regionID, err)
	}
	v.ActivatedAt = parseStamp(activated)
	v.UpdatedAt = parseStamp(updated)
	return v, nil
}
func (t *txStore) UpdateResponse(ctx context.Context, v domain.Response, expected int64) error {
	result, err := t.tx.ExecContext(ctx, `UPDATE responses SET level=?,status=?,activated_by=?,activated_at=?,reason=?,version=version+1,updated_at=? WHERE id=? AND version=?`, v.Level, v.Status, v.ActivatedBy, stamp(v.ActivatedAt), v.Reason, stamp(v.UpdatedAt), v.ID, expected)
	if err != nil {
		return translate("update", "response", v.ID, err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return domain.VersionConflict("response", v.ID, expected, v.Version)
	}
	return nil
}
