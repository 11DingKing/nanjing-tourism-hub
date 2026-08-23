package repository

import (
	"context"
	"database/sql"
	"time"
)

type DB interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func fmtTime(t time.Time) string     { return t.UTC().Format(time.RFC3339Nano) }
func parseTime(raw string) time.Time { t, _ := time.Parse(time.RFC3339Nano, raw); return t }
