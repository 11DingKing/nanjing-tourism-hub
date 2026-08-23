package sqlite

import (
	"context"
	"database/sql"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
)

func (t *txStore) InsertReceipt(ctx context.Context, v domain.Receipt) error {
	_, err := t.tx.ExecContext(ctx, `INSERT INTO receipts(id,task_id,incident_id,region_id,reporter_id,summary,status,evidence_count,submitted_at,reviewed_at,version) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, v.ID, v.TaskID, v.IncidentID, v.RegionID, v.ReporterID, v.Summary, v.Status, v.EvidenceCount, stamp(v.SubmittedAt), nullableStamp(v.ReviewedAt), v.Version)
	return translate("insert", "receipt", v.ID, err)
}
func (t *txStore) GetReceipt(ctx context.Context, id string) (domain.Receipt, error) {
	var v domain.Receipt
	var submitted string
	var reviewed sql.NullString
	err := t.tx.QueryRowContext(ctx, `SELECT id,task_id,incident_id,region_id,reporter_id,summary,status,evidence_count,submitted_at,reviewed_at,version FROM receipts WHERE id=?`, id).Scan(&v.ID, &v.TaskID, &v.IncidentID, &v.RegionID, &v.ReporterID, &v.Summary, &v.Status, &v.EvidenceCount, &submitted, &reviewed, &v.Version)
	if err != nil {
		return domain.Receipt{}, translate("get", "receipt", id, err)
	}
	v.SubmittedAt = parseStamp(submitted)
	v.ReviewedAt = scanOptional(reviewed)
	return v, nil
}
func (t *txStore) UpdateReceipt(ctx context.Context, v domain.Receipt, expected int64) error {
	result, err := t.tx.ExecContext(ctx, `UPDATE receipts SET summary=?,status=?,evidence_count=?,reviewed_at=?,version=version+1 WHERE id=? AND version=?`, v.Summary, v.Status, v.EvidenceCount, nullableStamp(v.ReviewedAt), v.ID, expected)
	if err != nil {
		return translate("update", "receipt", v.ID, err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return domain.VersionConflict("receipt", v.ID, expected, v.Version)
	}
	return nil
}
func (t *txStore) CountPendingReceipts(ctx context.Context, incidentID string) (int, error) {
	var count int
	err := t.tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM receipts WHERE incident_id=? AND status='submitted'`, incidentID).Scan(&count)
	return count, translate("count", "receipt", incidentID, err)
}
