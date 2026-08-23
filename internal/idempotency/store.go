package idempotency

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

type Store struct{ DB *sql.DB }

func Hash(body []byte) string { h := sha256.Sum256(body); return hex.EncodeToString(h[:]) }
func (s Store) Lookup(ctx context.Context, key, requestHash string) (string, bool, error) {
	var oldHash, response string
	e := s.DB.QueryRowContext(ctx, "SELECT request_hash,response FROM idempotency_keys WHERE key=?", key).Scan(&oldHash, &response)
	if e == sql.ErrNoRows {
		return "", false, nil
	}
	if e != nil {
		return "", false, e
	}
	if oldHash != requestHash {
		return "", false, fmt.Errorf("idempotency key reused with different request")
	}
	return response, true, nil
}
func (s Store) Save(ctx context.Context, key, requestHash, response string) error {
	_, e := s.DB.ExecContext(ctx, "INSERT INTO idempotency_keys(key,request_hash,response,created_at) VALUES(?,?,?,?)", key, requestHash, response, time.Now().UTC().Format(time.RFC3339Nano))
	return e
}
func (s Store) Prune(ctx context.Context, before time.Time) (int64, error) {
	r, e := s.DB.ExecContext(ctx, "DELETE FROM idempotency_keys WHERE created_at < ?", before.UTC().Format(time.RFC3339Nano))
	if e != nil {
		return 0, e
	}
	return r.RowsAffected()
}
