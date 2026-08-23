package compliance

import (
	"context"
	"testing"
	"time"
)

func TestEvidenceControlAndFinding(t *testing.T) {
	r := New()
	if e := r.AddControl(Control{ID: "c", Name: "field records", Required: true, EvidenceKinds: []string{"photo"}}); e != nil {
		t.Fatal(e)
	}
	if e := r.Collect(Evidence{ID: "e", ControlID: "c", ObjectID: "plot", ActorID: "u", Kind: "photo"}, []byte("photo")); e != nil {
		t.Fatal(e)
	}
	if e := r.OpenFinding(Finding{ID: "f", ControlID: "c", ObjectID: "plot", Severity: "high"}); e != nil {
		t.Fatal(e)
	}
	if e := r.AttachFinding("f", "e"); e != nil {
		t.Fatal(e)
	}
	if e := r.CloseFinding("f"); e != nil {
		t.Fatal(e)
	}
	if len(r.Due(time.Now())) != 0 {
		t.Fatal("closed finding due")
	}
	if e := Validate(context.Background(), Evidence{Hash: "x"}); e != nil {
		t.Fatal(e)
	}
}
func TestEvidenceExpiryAndScope(t *testing.T) {
	r := New()
	_ = r.AddControl(Control{ID: "c", Name: "x", EvidenceKinds: []string{"report"}})
	exp := time.Now().Add(-time.Hour)
	_ = r.Collect(Evidence{ID: "e", ControlID: "c", ObjectID: "o", ActorID: "u", Kind: "report", ExpiresAt: &exp}, []byte("x"))
	_ = r.OpenFinding(Finding{ID: "f", ControlID: "c", ObjectID: "o"})
	_ = r.AttachFinding("f", "e")
	if len(r.Due(time.Now())) != 1 {
		t.Fatal("expired evidence not due")
	}
	if e := r.AttachFinding("f", "missing"); e == nil {
		t.Fatal("missing evidence accepted")
	}
}
