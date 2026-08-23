package query

import (
	"strings"
	"testing"
)

func TestPagination(t *testing.T) {
	p, e := ParsePage("2", "25")
	if e != nil || p.Offset() != 25 {
		t.Fatalf("%+v %v", p, e)
	}
	if _, e = ParsePage("0", ""); e == nil {
		t.Fatal("zero page accepted")
	}
	if _, e = ParsePage("", "201"); e == nil {
		t.Fatal("oversized page accepted")
	}
}
func TestSQLFilterAndSort(t *testing.T) {
	sql, e := BuildList("safe_tasks", Filter{RegionID: "r1", Statuses: []string{"queued", "claimed"}}, Page{Number: 2, Size: 10}, Sort{Field: "-created_at"})
	if e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(sql.Text, "status IN (?,?)") || !strings.Contains(sql.Text, "created_at DESC") {
		t.Fatalf("%s", sql.Text)
	}
	if len(sql.Args) != 5 {
		t.Fatalf("args=%v", sql.Args)
	}
}
func TestSQLAllowlist(t *testing.T) {
	if _, _, e := CountSQL("tasks", Filter{}); e == nil {
		t.Fatal("unsafe table accepted")
	}
}
