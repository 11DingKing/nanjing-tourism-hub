package recovery

import (
	"context"
	"testing"
	"time"
)

func TestPlanDependenciesAndEvidence(t *testing.T) {
	now := time.Now()
	p := NewPlan("p", "r", "d", []Step{{ID: "drain", Name: "drain", State: Pending, DueAt: now.Add(time.Hour)}, {ID: "repair", Name: "repair", Dependency: "drain", State: Pending, DueAt: now.Add(time.Hour)}})
	if e := p.Assign("repair", "crew"); e != nil {
		t.Fatal(e)
	}
	if e := p.Start("repair"); e == nil {
		t.Fatal("dependency bypass")
	}
	if e := p.Assign("drain", "crew"); e != nil {
		t.Fatal(e)
	}
	if e := Run(context.Background(), p, "drain"); e != nil {
		t.Fatal(e)
	}
	if e := p.Start("repair"); e != nil {
		t.Fatal(e)
	}
	if e := p.Verify("repair", []string{"photo", "receipt"}); e != nil {
		t.Fatal(e)
	}
	if !p.Complete() {
		t.Fatal("plan incomplete")
	}
}
func TestPlanCopyAndCancellation(t *testing.T) {
	p := NewPlan("p", "r", "d", []Step{{ID: "x", State: Pending}})
	if e := p.Assign("x", "a"); e != nil {
		t.Fatal(e)
	}
	if e := p.Cancel("x"); e != nil {
		t.Fatal(e)
	}
	s := p.Snapshot()
	s[0].Evidence = append(s[0].Evidence, "mutated")
	if len(p.Snapshot()[0].Evidence) != 0 {
		t.Fatal("snapshot alias")
	}
}
