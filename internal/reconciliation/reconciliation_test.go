package reconciliation

import "testing"

func TestReconciliation(t *testing.T) {
	b := New()
	if e := b.Record(Entry{ID: "a", RegionID: "r", Reference: "aid", Expected: 100, Actual: 101}); e != nil {
		t.Fatal(e)
	}
	if got := b.Reconcile("a", 2); !got.Matched {
		t.Fatal(got)
	}
	if b.Balance("r") != 1 {
		t.Fatal("balance")
	}
	if e := b.Record(Entry{ID: "b", RegionID: "r", Reference: "aid2", Expected: 100, Actual: 120}); e != nil {
		t.Fatal(e)
	}
	if got := b.Reconcile("b", 2); got.Matched {
		t.Fatal(got)
	}
	if len(b.Exceptions("r")) != 1 {
		t.Fatal("exceptions")
	}
}
func TestReconciliationBatch(t *testing.T) {
	b := New()
	_ = b.Record(Entry{ID: "2", RegionID: "r", Reference: "b", Expected: 2, Actual: 2})
	_ = b.Record(Entry{ID: "1", RegionID: "r", Reference: "a", Expected: 3, Actual: 1})
	r := b.Batch([]string{"2", "1"}, 0)
	if len(r) != 2 || r[0].EntryID != "1" {
		t.Fatal(r)
	}
	if e := Validate(Entry{Expected: -1}); e == nil {
		t.Fatal("negative accepted")
	}
}
