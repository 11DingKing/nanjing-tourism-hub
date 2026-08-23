package sqlite

import (
	"context"
	"database/sql"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
)

func (t *txStore) GetTeam(ctx context.Context, id string) (domain.Team, error) {
	var v domain.Team
	var incident sql.NullString
	var updated string
	err := t.tx.QueryRowContext(ctx, `SELECT id,region_id,name,specialty,status,current_incident_id,leader_id,version,updated_at FROM teams WHERE id=?`, id).Scan(&v.ID, &v.RegionID, &v.Name, &v.Specialty, &v.Status, &incident, &v.LeaderID, &v.Version, &updated)
	if err != nil {
		return domain.Team{}, translate("get", "team", id, err)
	}
	v.CurrentIncidentID = incident.String
	v.UpdatedAt = parseStamp(updated)
	return v, nil
}
func (t *txStore) UpdateTeam(ctx context.Context, v domain.Team, expected int64) error {
	result, err := t.tx.ExecContext(ctx, `UPDATE teams SET status=?,current_incident_id=?,leader_id=?,version=version+1,updated_at=? WHERE id=? AND version=?`, v.Status, nullText(v.CurrentIncidentID), v.LeaderID, stamp(v.UpdatedAt), v.ID, expected)
	if err != nil {
		return translate("update", "team", v.ID, err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return domain.VersionConflict("team", v.ID, expected, v.Version)
	}
	return nil
}
func (t *txStore) InsertDispatch(ctx context.Context, v domain.Dispatch) error {
	_, err := t.tx.ExecContext(ctx, `INSERT INTO dispatches(id,team_id,incident_id,region_id,task_id,status,requested_by,accepted_at,completed_at,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, v.ID, v.TeamID, v.IncidentID, v.RegionID, v.TaskID, v.Status, v.RequestedBy, nullableStamp(v.AcceptedAt), nullableStamp(v.CompletedAt), v.Version, stamp(v.CreatedAt), stamp(v.UpdatedAt))
	return translate("insert", "dispatch", v.ID, err)
}
func (t *txStore) GetDispatch(ctx context.Context, id string) (domain.Dispatch, error) {
	var v domain.Dispatch
	var accepted, completed sql.NullString
	var created, updated string
	err := t.tx.QueryRowContext(ctx, `SELECT id,team_id,incident_id,region_id,task_id,status,requested_by,accepted_at,completed_at,version,created_at,updated_at FROM dispatches WHERE id=?`, id).Scan(&v.ID, &v.TeamID, &v.IncidentID, &v.RegionID, &v.TaskID, &v.Status, &v.RequestedBy, &accepted, &completed, &v.Version, &created, &updated)
	if err != nil {
		return domain.Dispatch{}, translate("get", "dispatch", id, err)
	}
	v.AcceptedAt = scanOptional(accepted)
	v.CompletedAt = scanOptional(completed)
	v.CreatedAt = parseStamp(created)
	v.UpdatedAt = parseStamp(updated)
	return v, nil
}
func (t *txStore) UpdateDispatch(ctx context.Context, v domain.Dispatch, expected int64) error {
	result, err := t.tx.ExecContext(ctx, `UPDATE dispatches SET team_id=?,status=?,accepted_at=?,completed_at=?,version=version+1,updated_at=? WHERE id=? AND version=?`, v.TeamID, v.Status, nullableStamp(v.AcceptedAt), nullableStamp(v.CompletedAt), stamp(v.UpdatedAt), v.ID, expected)
	if err != nil {
		return translate("update", "dispatch", v.ID, err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return domain.VersionConflict("dispatch", v.ID, expected, v.Version)
	}
	return nil
}
