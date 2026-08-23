package repository

import (
	"context"
	"database/sql"
	"github.com/11DingKing/autumn-grain-resilience/internal/domain"
)

type AuditRepo struct{ DB DB }

func (r AuditRepo) Append(ctx context.Context, v domain.AuditEvent) error {
	_, e := r.DB.ExecContext(ctx, "INSERT INTO audit_events(id,actor_id,action,object_type,object_id,result,request_id,created_at) VALUES(?,?,?,?,?,?,?,?)", v.ID, v.ActorID, v.Action, v.ObjectType, v.ObjectID, v.Result, v.RequestID, fmtTime(v.CreatedAt))
	return e
}
func (r AuditRepo) AppendTx(ctx context.Context, tx *sql.Tx, v domain.AuditEvent) error {
	_, e := tx.ExecContext(ctx, "INSERT INTO audit_events(id,actor_id,action,object_type,object_id,result,request_id,created_at) VALUES(?,?,?,?,?,?,?,?)", v.ID, v.ActorID, v.Action, v.ObjectType, v.ObjectID, v.Result, v.RequestID, fmtTime(v.CreatedAt))
	return e
}
