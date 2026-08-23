package repository

import (
	"context"
	"database/sql"
	"github.com/11DingKing/autumn-grain-resilience/internal/domain"
	"time"
)

type OutboxRepo struct{ DB DB }

func (r OutboxRepo) Add(ctx context.Context, v domain.OutboxEvent) error {
	_, e := r.DB.ExecContext(ctx, "INSERT INTO outbox_events(id,topic,aggregate_id,payload,status,attempts,available_at) VALUES(?,?,?,?,?,?,?)", v.ID, v.Topic, v.AggregateID, v.Payload, v.Status, v.Attempts, fmtTime(v.AvailableAt))
	return e
}
func (r OutboxRepo) Ready(ctx context.Context, now time.Time, limit int) ([]domain.OutboxEvent, error) {
	rows, e := r.DB.QueryContext(ctx, "SELECT id,topic,aggregate_id,payload,status,attempts,available_at FROM outbox_events WHERE status='pending' AND available_at<=? ORDER BY available_at LIMIT ?", fmtTime(now), limit)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []domain.OutboxEvent
	for rows.Next() {
		var v domain.OutboxEvent
		var at string
		if e = rows.Scan(&v.ID, &v.Topic, &v.AggregateID, &v.Payload, &v.Status, &v.Attempts, &at); e != nil {
			return nil, e
		}
		v.AvailableAt = parseTime(at)
		out = append(out, v)
	}
	return out, rows.Err()
}
func (r OutboxRepo) Mark(ctx context.Context, id, status string, attempts int, next time.Time) error {
	_, e := r.DB.ExecContext(ctx, "UPDATE outbox_events SET status=?,attempts=?,available_at=? WHERE id=?", status, attempts, fmtTime(next), id)
	return e
}

var _ DB = (*sql.DB)(nil)
