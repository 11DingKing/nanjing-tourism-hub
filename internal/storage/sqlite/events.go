package sqlite

import (
	"context"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
)

func (t *txStore) InsertAudit(ctx context.Context, v domain.AuditEvent) error {
	_, err := t.tx.ExecContext(ctx, `INSERT INTO audit_events(id,actor_id,request_id,action,entity,entity_id,result,detail,created_at) VALUES(?,?,?,?,?,?,?,?,?)`, v.ID, v.ActorID, v.RequestID, v.Action, v.Entity, v.EntityID, v.Result, v.Detail, stamp(v.CreatedAt))
	return translate("insert", "audit_event", v.ID, err)
}
func (t *txStore) InsertOutbox(ctx context.Context, v domain.OutboxEvent) error {
	_, err := t.tx.ExecContext(ctx, `INSERT INTO outbox_events(id,topic,aggregate_type,aggregate_id,payload,status,attempts,next_attempt_at,last_error,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, v.ID, v.Topic, v.AggregateType, v.AggregateID, v.Payload, v.Status, v.Attempts, stamp(v.NextAttemptAt), v.LastError, stamp(v.CreatedAt), stamp(v.UpdatedAt))
	return translate("insert", "outbox_event", v.ID, err)
}
func (t *txStore) InsertJob(ctx context.Context, v domain.Job) error {
	_, err := t.tx.ExecContext(ctx, `INSERT INTO jobs(id,kind,aggregate_id,payload,status,attempts,max_attempts,next_run_at,last_error,locked_until,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, v.ID, v.Kind, v.AggregateID, v.Payload, v.Status, v.Attempts, v.MaxAttempts, stamp(v.NextRunAt), v.LastError, nullableStamp(v.LockedUntil), stamp(v.CreatedAt), stamp(v.UpdatedAt))
	return translate("insert", "job", v.ID, err)
}
func (t *txStore) GetIdempotency(ctx context.Context, scope, key string) (domain.IdempotencyRecord, error) {
	var v domain.IdempotencyRecord
	var expires, created string
	err := t.tx.QueryRowContext(ctx, `SELECT scope,key,request_hash,status_code,response,expires_at,created_at FROM idempotency_records WHERE scope=? AND key=?`, scope, key).Scan(&v.Scope, &v.Key, &v.RequestHash, &v.StatusCode, &v.Response, &expires, &created)
	if err != nil {
		return domain.IdempotencyRecord{}, translate("get", "idempotency", scope+":"+key, err)
	}
	v.ExpiresAt = parseStamp(expires)
	v.CreatedAt = parseStamp(created)
	return v, nil
}
func (t *txStore) PutIdempotency(ctx context.Context, v domain.IdempotencyRecord) error {
	_, err := t.tx.ExecContext(ctx, `INSERT INTO idempotency_records(scope,key,request_hash,status_code,response,expires_at,created_at) VALUES(?,?,?,?,?,?,?)`, v.Scope, v.Key, v.RequestHash, v.StatusCode, v.Response, stamp(v.ExpiresAt), stamp(v.CreatedAt))
	return translate("put", "idempotency", v.Scope+":"+v.Key, err)
}
