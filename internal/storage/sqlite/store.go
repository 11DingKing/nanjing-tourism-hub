package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
	"github.com/11DingKing/nanjing-tourism-hub/internal/repository"
	_ "modernc.org/sqlite"
)

type Store struct{ db *sql.DB }
type txStore struct{ tx *sql.Tx }

func Open(path string) (*Store, error) {
	dsn := path
	if path != ":memory:" && path[:min(len(path), 5)] != "file:" {
		dsn = "file:" + path
	}
	db, err := sql.Open("sqlite", dsn+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(4)
	db.SetConnMaxLifetime(time.Hour)
	store := &Store{db: db}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := store.Ping(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error { return s.db.Close() }
func (s *Store) Ping(ctx context.Context) error {
	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping sqlite: %w", err)
	}
	return nil
}
func (s *Store) Readiness(ctx context.Context) error {
	var value int
	if err := s.db.QueryRowContext(ctx, "SELECT 1").Scan(&value); err != nil {
		return fmt.Errorf("readiness: %w", err)
	}
	if value != 1 {
		return fmt.Errorf("readiness: unexpected probe")
	}
	return nil
}

func (s *Store) WithinTx(ctx context.Context, fn func(repository.Tx) error) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	if err := fn(&txStore{tx: tx}); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	committed = true
	return nil
}

func translate(op, entity, id string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Wrap(domain.ErrNotFound, op, entity, id, "record was not found", err)
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return domain.Wrap(domain.ErrCanceled, op, entity, id, "database operation canceled", err)
	}
	if contains(err.Error(), "UNIQUE constraint failed") {
		return domain.Wrap(domain.ErrConflict, op, entity, id, "unique constraint", err)
	}
	if contains(err.Error(), "FOREIGN KEY constraint failed") {
		return domain.Wrap(domain.ErrInvalidState, op, entity, id, "related record is missing", err)
	}
	return domain.Wrap(domain.ErrDependency, op, entity, id, "database operation failed", err)
}

func contains(value, target string) bool {
	if len(target) == 0 {
		return true
	}
	for i := 0; i+len(target) <= len(value); i++ {
		if value[i:i+len(target)] == target {
			return true
		}
	}
	return false
}
