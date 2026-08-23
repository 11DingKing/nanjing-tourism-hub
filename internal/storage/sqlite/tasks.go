package sqlite

import (
	"context"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
)

func (t *txStore) InsertTask(ctx context.Context, v domain.DistrictTask) error {
	_, err := t.tx.ExecContext(ctx, `INSERT INTO district_tasks(id,incident_id,region_id,kind,summary,status,priority,assignee_id,created_by,deadline,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`, v.ID, v.IncidentID, v.RegionID, v.Kind, v.Summary, v.Status, v.Priority, nullText(v.AssigneeID), v.CreatedBy, stamp(v.Deadline), v.Version, stamp(v.CreatedAt), stamp(v.UpdatedAt))
	return translate("insert", "task", v.ID, err)
}
func (t *txStore) GetTask(ctx context.Context, id string) (domain.DistrictTask, error) {
	var v domain.DistrictTask
	var assignee *string
	var deadline, created, updated string
	err := t.tx.QueryRowContext(ctx, `SELECT id,incident_id,region_id,kind,summary,status,priority,assignee_id,created_by,deadline,version,created_at,updated_at FROM district_tasks WHERE id=?`, id).Scan(&v.ID, &v.IncidentID, &v.RegionID, &v.Kind, &v.Summary, &v.Status, &v.Priority, &assignee, &v.CreatedBy, &deadline, &v.Version, &created, &updated)
	if err != nil {
		return domain.DistrictTask{}, translate("get", "task", id, err)
	}
	if assignee != nil {
		v.AssigneeID = *assignee
	}
	v.Deadline = parseStamp(deadline)
	v.CreatedAt = parseStamp(created)
	v.UpdatedAt = parseStamp(updated)
	return v, nil
}
func (t *txStore) UpdateTask(ctx context.Context, v domain.DistrictTask, expected int64) error {
	result, err := t.tx.ExecContext(ctx, `UPDATE district_tasks SET status=?,priority=?,assignee_id=?,deadline=?,version=version+1,updated_at=? WHERE id=? AND version=?`, v.Status, v.Priority, nullText(v.AssigneeID), stamp(v.Deadline), stamp(v.UpdatedAt), v.ID, expected)
	if err != nil {
		return translate("update", "task", v.ID, err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return domain.VersionConflict("task", v.ID, expected, v.Version)
	}
	return nil
}
func (t *txStore) CountOpenTasks(ctx context.Context, incidentID string) (int, error) {
	var count int
	err := t.tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM district_tasks WHERE incident_id=? AND status NOT IN ('completed','canceled')`, incidentID).Scan(&count)
	return count, translate("count", "task", incidentID, err)
}
