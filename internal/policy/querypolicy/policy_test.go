package querypolicy

import (
	"context"
	"testing"
	"time"
)

func validInput() Input {
	now := time.Date(2026, 8, 23, 8, 0, 0, 0, time.UTC)
	return Input{ActorRole: "duty_officer", ActorRegion: "", ObjectRegion: "", CurrentState: "requested", TargetState: "authorized", Capacity: 100, Committed: 20, Requested: 10, ExpectedVersion: 4, ActualVersion: 4, DependencyCount: 0, Now: now, Deadline: now.Add(time.Hour)}
}

func TestPolicyAllowsValidOperation(t *testing.T) {
	decision := Evaluate(context.Background(), validInput())
	if !decision.Allowed {
		t.Fatalf("expected operation to be allowed: %#v", decision)
	}
	if decision.Code != "allowed" {
		t.Fatalf("unexpected code %q", decision.Code)
	}
	if len(decision.Obligations) != 3 {
		t.Fatalf("unexpected obligations %#v", decision.Obligations)
	}
}

func TestPolicyRejectsInvalidConditions(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Input)
		code   string
	}{
		{name: "role", mutate: func(v *Input) { v.ActorRole = "visitor" }, code: "role_forbidden"},
		{name: "state", mutate: func(v *Input) { v.TargetState = "unknown" }, code: "transition_rejected"},
		{name: "version", mutate: func(v *Input) { v.ActualVersion = 8 }, code: "version_conflict"},
		{name: "negative", mutate: func(v *Input) { v.Requested = -1 }, code: "negative_quantity"},
		{name: "capacity", mutate: func(v *Input) { v.Requested = 90 }, code: "capacity_exceeded"},
		{name: "deadline", mutate: func(v *Input) { v.Deadline = v.Now }, code: "deadline_expired"},
		{name: "dependency", mutate: func(v *Input) { v.DependencyCount = -1 }, code: "dependency_unknown"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := validInput()
			test.mutate(&input)
			decision := Evaluate(context.Background(), input)
			if decision.Allowed {
				t.Fatalf("expected rejection for %s", test.name)
			}
			if decision.Code != test.code {
				t.Fatalf("got %q want %q", decision.Code, test.code)
			}
			if RequireAllowed(decision) == nil {
				t.Fatal("expected policy error")
			}
		})
	}
}

func TestPolicyHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	decision := Evaluate(ctx, validInput())
	if decision.Code != "request_canceled" {
		t.Fatalf("unexpected decision %#v", decision)
	}
}

func TestExplanationCopiesMutableValues(t *testing.T) {
	decision := Evaluate(context.Background(), validInput())
	metadata := map[string]string{" region ": " coast-a "}
	explanation := Explain(decision, metadata)
	metadata["region"] = "changed"
	decision.Obligations[0] = "changed"
	if explanation.Metadata["region"] != "coast-a" {
		t.Fatalf("metadata was not normalized: %#v", explanation.Metadata)
	}
	if explanation.Obligations[0] == "changed" {
		t.Fatal("explanation shared obligation storage")
	}
}

func TestMergePreservesUniqueObligations(t *testing.T) {
	left := Decision{Allowed: true, Code: "allowed", Obligations: []string{"audit", "persist"}}
	right := Decision{Allowed: true, Code: "allowed", Obligations: []string{"persist", "publish"}}
	merged := Merge(left, right)
	set := ObligationSet(merged)
	for _, expected := range []string{"audit", "persist", "publish"} {
		if !set[expected] {
			t.Fatalf("missing %s in %#v", expected, merged.Obligations)
		}
	}
}
