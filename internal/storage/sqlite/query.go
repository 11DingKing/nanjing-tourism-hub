package sqlite

import (
	"context"
	"fmt"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
	"github.com/11DingKing/nanjing-tourism-hub/internal/repository"
	"strings"
)

func (s *Store) ListTasks(ctx context.Context, filter domain.TaskFilter) (domain.TaskPage, error) {
	page := domain.ValidatePage(filter.Page)
	clauses := []string{"1=1"}
	args := []any{}
	if filter.IncidentID != "" {
		clauses = append(clauses, "incident_id=?")
		args = append(args, filter.IncidentID)
	}
	if filter.RegionID != "" {
		clauses = append(clauses, "region_id=?")
		args = append(args, filter.RegionID)
	}
	if filter.Status != "" {
		clauses = append(clauses, "status=?")
		args = append(args, filter.Status)
	}
	if filter.Kind != "" {
		clauses = append(clauses, "kind=?")
		args = append(args, filter.Kind)
	}
	if filter.DeadlineBefore != nil {
		clauses = append(clauses, "deadline<?")
		args = append(args, stamp(*filter.DeadlineBefore))
	}
	where := strings.Join(clauses, " AND ")
	var total int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM district_tasks WHERE "+where, args...).Scan(&total); err != nil {
		return domain.TaskPage{}, translate("count", "task", "", err)
	}
	sort := "deadline"
	switch page.Sort {
	case "priority":
		sort = "priority"
	case "created_at":
		sort = "created_at"
	}
	query := fmt.Sprintf("SELECT id,incident_id,region_id,kind,summary,status,priority,COALESCE(assignee_id,''),created_by,deadline,version,created_at,updated_at FROM district_tasks WHERE %s ORDER BY %s %s,id ASC LIMIT ? OFFSET ?", where, sort, page.Direction)
	args = append(args, page.Limit, page.Offset)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return domain.TaskPage{}, translate("list", "task", "", err)
	}
	defer rows.Close()
	items := []domain.DistrictTask{}
	for rows.Next() {
		var v domain.DistrictTask
		var deadline, created, updated string
		if err := rows.Scan(&v.ID, &v.IncidentID, &v.RegionID, &v.Kind, &v.Summary, &v.Status, &v.Priority, &v.AssigneeID, &v.CreatedBy, &deadline, &v.Version, &created, &updated); err != nil {
			return domain.TaskPage{}, err
		}
		v.Deadline = parseStamp(deadline)
		v.CreatedAt = parseStamp(created)
		v.UpdatedAt = parseStamp(updated)
		items = append(items, v)
	}
	return domain.TaskPage{Items: items, Total: total}, rows.Err()
}
func (s *Store) SnapshotIncident(ctx context.Context, id string) (repository.IncidentSnapshot, error) {
	snapshot := repository.IncidentSnapshot{}
	err := s.WithinTx(ctx, func(tx repository.Tx) error {
		incident, err := tx.GetIncident(ctx, id)
		if err != nil {
			return err
		}
		snapshot.Incident = incident
		return nil
	})
	return snapshot, err
}
