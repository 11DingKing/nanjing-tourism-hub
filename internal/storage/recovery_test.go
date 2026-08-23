package storage

import (
	"context"
	"testing"
)

func TestDatabaseReopenPersistsMigrationAndData(t *testing.T) {
	path := t.TempDir() + "/state.db"
	db, e := Open(path)
	if e != nil {
		t.Fatal(e)
	}
	if e = Migrate(context.Background(), db); e != nil {
		t.Fatal(e)
	}
	if _, e = db.Exec("INSERT INTO regions(id,name,flood_risk) VALUES('r','North','high')"); e != nil {
		t.Fatal(e)
	}
	if e = db.Close(); e != nil {
		t.Fatal(e)
	}
	db, e = Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	if e = Migrate(context.Background(), db); e != nil {
		t.Fatal(e)
	}
	var name string
	if e = db.QueryRow("SELECT name FROM regions WHERE id='r'").Scan(&name); e != nil || name != "North" {
		t.Fatalf("%s %v", name, e)
	}
}
