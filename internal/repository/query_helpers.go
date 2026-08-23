package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type QueryPage struct{ Limit, Offset int }

func ListArgs(page QueryPage) (string, []any) {
	if page.Limit < 1 || page.Limit > 200 {
		page.Limit = 50
	}
	if page.Offset < 0 {
		page.Offset = 0
	}
	return " LIMIT ? OFFSET ?", []any{page.Limit, page.Offset}
}
func ScanStrings(rows *sql.Rows) ([][]string, error) {
	defer rows.Close()
	cols, e := rows.Columns()
	if e != nil {
		return nil, e
	}
	out := [][]string{}
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if e = rows.Scan(ptrs...); e != nil {
			return nil, e
		}
		row := make([]string, len(cols))
		for i, v := range vals {
			row[i] = fmt.Sprint(v)
		}
		out = append(out, row)
	}
	return out, rows.Err()
}
func ExecInBatches(ctx context.Context, db DB, query string, items []string, batch int) error {
	if batch < 1 {
		batch = 50
	}
	for start := 0; start < len(items); start += batch {
		end := start + batch
		if end > len(items) {
			end = len(items)
		}
		args := make([]any, end-start)
		for i := range args {
			args[i] = items[start+i]
		}
		marks := make([]string, len(args))
		for i := range marks {
			marks[i] = "?"
		}
		if _, e := db.ExecContext(ctx, strings.Replace(query, "?", strings.Join(marks, ","), 1), args...); e != nil {
			return e
		}
	}
	return nil
}
func RetryDelay(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	if attempt > 8 {
		attempt = 8
	}
	return time.Duration(1<<attempt) * 250 * time.Millisecond
}
