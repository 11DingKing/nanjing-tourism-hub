package sqlite

import (
	"context"
	"database/sql"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
	"time"
)

func (s *Store) ClaimDueJobs(ctx context.Context, now time.Time, limit int, lease time.Duration) ([]domain.Job, error) {
	if limit < 1 {
		return []domain.Job{}, nil
	}
	result := make([]domain.Job, 0, limit)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, translate("claim", "job", "", err)
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, `SELECT id,kind,aggregate_id,payload,status,attempts,max_attempts,next_run_at,last_error,locked_until,created_at,updated_at FROM jobs WHERE status IN ('pending','retry') AND next_run_at<=? AND (locked_until IS NULL OR locked_until<?) ORDER BY next_run_at,id LIMIT ?`, stamp(now), stamp(now), limit)
	if err != nil {
		return nil, translate("list", "job", "", err)
	}
	for rows.Next() {
		var v domain.Job
		var next, created, updated string
		var locked sql.NullString
		if err := rows.Scan(&v.ID, &v.Kind, &v.AggregateID, &v.Payload, &v.Status, &v.Attempts, &v.MaxAttempts, &next, &v.LastError, &locked, &created, &updated); err != nil {
			rows.Close()
			return nil, translate("scan", "job", "", err)
		}
		v.NextRunAt = parseStamp(next)
		v.LockedUntil = scanOptional(locked)
		v.CreatedAt = parseStamp(created)
		v.UpdatedAt = parseStamp(updated)
		result = append(result, v)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	until := now.Add(lease)
	for _, v := range result {
		if _, err := tx.ExecContext(ctx, `UPDATE jobs SET status='running',locked_until=?,updated_at=? WHERE id=?`, stamp(until), stamp(now), v.ID); err != nil {
			return nil, translate("lock", "job", v.ID, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, translate("commit", "job", "", err)
	}
	return result, nil
}
func (s *Store) ClaimOutbox(ctx context.Context, now time.Time, limit int) ([]domain.OutboxEvent, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,topic,aggregate_type,aggregate_id,payload,status,attempts,next_attempt_at,last_error,created_at,updated_at FROM outbox_events WHERE status IN ('pending','retry') AND next_attempt_at<=? ORDER BY next_attempt_at,id LIMIT ?`, stamp(now), limit)
	if err != nil {
		return nil, translate("list", "outbox_event", "", err)
	}
	defer rows.Close()
	items := []domain.OutboxEvent{}
	for rows.Next() {
		var v domain.OutboxEvent
		var next, created, updated string
		if err := rows.Scan(&v.ID, &v.Topic, &v.AggregateType, &v.AggregateID, &v.Payload, &v.Status, &v.Attempts, &next, &v.LastError, &created, &updated); err != nil {
			return nil, err
		}
		v.NextAttemptAt = parseStamp(next)
		v.CreatedAt = parseStamp(created)
		v.UpdatedAt = parseStamp(updated)
		items = append(items, v)
	}
	return items, rows.Err()
}
func (s *Store) CompleteJob(ctx context.Context, id string, at time.Time) error {
	result, err := s.db.ExecContext(ctx, `UPDATE jobs SET status='completed',locked_until=NULL,last_error='',updated_at=? WHERE id=? AND status='running'`, stamp(at), id)
	if err != nil {
		return translate("complete", "job", id, err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return domain.Wrap(domain.ErrConflict, "complete", "job", id, "job is not running", nil)
	}
	return nil
}
func (s *Store) FailJob(ctx context.Context, id, message string, next time.Time) error {
	result, err := s.db.ExecContext(ctx, `UPDATE jobs SET status=CASE WHEN attempts+1>=max_attempts THEN 'dead' ELSE 'retry' END,attempts=attempts+1,locked_until=NULL,last_error=?,next_run_at=?,updated_at=? WHERE id=? AND status='running'`, message, stamp(next), stamp(time.Now().UTC()), id)
	if err != nil {
		return translate("fail", "job", id, err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return domain.Wrap(domain.ErrConflict, "fail", "job", id, "job is not running", nil)
	}
	return nil
}
func (s *Store) CompleteOutbox(ctx context.Context, id string, at time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE outbox_events SET status='published',last_error='',updated_at=? WHERE id=? AND status IN ('pending','retry')`, stamp(at), id)
	return translate("complete", "outbox_event", id, err)
}
func (s *Store) FailOutbox(ctx context.Context, id, message string, next time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE outbox_events SET status='retry',attempts=attempts+1,last_error=?,next_attempt_at=?,updated_at=? WHERE id=? AND status IN ('pending','retry')`, message, stamp(next), stamp(time.Now().UTC()), id)
	return translate("fail", "outbox_event", id, err)
}
