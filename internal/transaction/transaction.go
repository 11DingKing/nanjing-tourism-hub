package transaction

import (
	"context"
	"database/sql"
	"fmt"
)

type Manager struct{ DB *sql.DB }

func (m Manager) Run(ctx context.Context, fn func(context.Context, *sql.Tx) error) error {
	tx, e := m.DB.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	if e = fn(ctx, tx); e != nil {
		return e
	}
	if e = tx.Commit(); e != nil {
		return fmt.Errorf("commit transaction: %w", e)
	}
	return nil
}
func WithSavepoint(ctx context.Context, tx *sql.Tx, name string, fn func() error) error {
	if _, e := tx.ExecContext(ctx, "SAVEPOINT "+name); e != nil {
		return e
	}
	if e := fn(); e != nil {
		_, _ = tx.ExecContext(ctx, "ROLLBACK TO "+name)
		return e
	}
	_, e := tx.ExecContext(ctx, "RELEASE "+name)
	return e
}
func RollbackOnly(_ context.Context, tx *sql.Tx) error { return tx.Rollback() }
