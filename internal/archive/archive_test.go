package archive

import "testing"

func TestArchiveRoundTrip(t *testing.T) {
	s := New()
	if e := s.Put(Record{ID: "r", Kind: "observation", RegionID: "x", Payload: map[string]any{"severity": "high"}}); e != nil {
		t.Fatal(e)
	}
	if e := s.Archive("r"); e != nil {
		t.Fatal(e)
	}
	r, e := s.Restore("r")
	if e != nil {
		t.Fatal(e)
	}
	if r.Payload == nil || r.Checksum == "" {
		t.Fatal("restore")
	}
	if e = s.Archive("r"); e != nil {
		t.Fatal(e)
	}
}
