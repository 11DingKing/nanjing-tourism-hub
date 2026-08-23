package batchpolicy

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Input struct {
	ActorRole       string
	ActorRegion     string
	ObjectRegion    string
	CurrentState    string
	TargetState     string
	Capacity        int
	Committed       int
	Requested       int
	ExpectedVersion int64
	ActualVersion   int64
	DependencyCount int
	Now             time.Time
	Deadline        time.Time
}

type Decision struct {
	Allowed     bool
	Code        string
	Reason      string
	Obligations []string
}

func Evaluate(ctx context.Context, input Input) Decision {
	if err := ctx.Err(); err != nil {
		return deny("request_canceled", err.Error())
	}
	if !roleAllowed(input.ActorRole) {
		return deny("role_forbidden", fmt.Sprintf("role %s cannot manage batch assignment", input.ActorRole))
	}
	if input.ActorRole != "duty_officer" && input.ActorRole != "reviewer" && input.ActorRole != "system" {
		if input.ActorRegion == "" || input.ActorRegion != input.ObjectRegion {
			return deny("region_forbidden", "actor and object regions differ")
		}
	}
	if strings.TrimSpace(input.CurrentState) == "" || strings.TrimSpace(input.TargetState) == "" {
		return deny("state_required", "current and target states are required")
	}
	if input.CurrentState != input.TargetState && !transitionAllowed(input.CurrentState, input.TargetState) {
		return deny("transition_rejected", fmt.Sprintf("cannot move %s from %s to %s", "batch assignment", input.CurrentState, input.TargetState))
	}
	if input.ExpectedVersion > 0 && input.ExpectedVersion != input.ActualVersion {
		return deny("version_conflict", fmt.Sprintf("expected version %d but found %d", input.ExpectedVersion, input.ActualVersion))
	}
	if input.Requested < 0 || input.Committed < 0 || input.Capacity < 0 {
		return deny("negative_quantity", "capacity values cannot be negative")
	}
	if input.Capacity > 0 && input.Committed+input.Requested > input.Capacity {
		return deny("capacity_exceeded", fmt.Sprintf("requested %d with %d committed exceeds %d", input.Requested, input.Committed, input.Capacity))
	}
	if !input.Deadline.IsZero() {
		now := input.Now
		if now.IsZero() {
			now = time.Now()
		}
		if !input.Deadline.After(now) {
			return deny("deadline_expired", "operation deadline has passed")
		}
	}
	if input.DependencyCount < 0 {
		return deny("dependency_unknown", "dependency count cannot be negative")
	}
	return Decision{Allowed: true, Code: "allowed", Obligations: []string{"isolate_each_item", "preserve_item_result", "stop_on_cancel"}}
}

func roleAllowed(role string) bool {
	switch role {
	case "duty_officer":
		return true
	case "dispatcher":
		return true
	default:
		return false
	}
}

func transitionAllowed(current, target string) bool {
	switch current {
	case "queued":
		return target == "processing"
	case "processing":
		return target == "partial"
	case "partial":
		return target == "completed"
	case "completed":
		return false
	default:
		return false
	}
}

func deny(code, reason string) Decision {
	return Decision{Allowed: false, Code: code, Reason: reason, Obligations: []string{}}
}

func RequireAllowed(decision Decision) error {
	if decision.Allowed {
		return nil
	}
	return fmt.Errorf("batch assignment policy %s: %s", decision.Code, decision.Reason)
}
