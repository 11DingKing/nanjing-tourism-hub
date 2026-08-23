package transaction

import (
	"context"
	"database/sql"
	"errors"
	"github.com/11DingKing/autumn-grain-resilience/internal/storage"
	"testing"
)

func TestRunCommitsAndRollsBack(t *testing.T) {
	db, e := storage.Open(":memory:")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	m := Manager{DB: db}
	if e = m.Run(context.Background(), func(_ context.Context, tx *sql.Tx) error { _, e := tx.Exec("CREATE TABLE x(v TEXT)"); return e }); e != nil {
		t.Fatal(e)
	}
	if e = m.Run(context.Background(), func(_ context.Context, tx *sql.Tx) error {
		_, _ = tx.Exec("INSERT INTO x(v) VALUES('bad')")
		return errors.New("fail")
	}); e == nil {
		t.Fatal("expected rollback")
	}
	var n int
	if e = db.QueryRow("SELECT COUNT(*) FROM x").Scan(&n); e != nil {
		t.Fatal(e)
	}
	if n != 0 {
		t.Fatal("rollback failed")
	}
}
