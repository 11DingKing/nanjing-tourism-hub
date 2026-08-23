package policy

import (
	"fmt"
	"strings"
	"time"
)

type Decision struct {
	Allowed bool
	Code    string
	Reason  string
}
type Principal struct {
	ID      string
	Role    string
	Regions map[string]bool
}
type Resource struct {
	RegionID  string
	OwnerID   string
	State     string
	UpdatedAt time.Time
}

func Allow() Decision                   { return Decision{Allowed: true, Code: "allowed"} }
func Deny(code, reason string) Decision { return Decision{Code: code, Reason: reason} }
func CanRead(p Principal, r Resource) Decision {
	if p.ID == "" {
		return Deny("unauthenticated", "principal is missing")
	}
	if p.Role == "admin" || p.Role == "auditor" {
		return Allow()
	}
	if !p.Regions[r.RegionID] {
		return Deny("region_forbidden", "resource belongs to another region")
	}
	return Allow()
}
func CanMutate(p Principal, r Resource) Decision {
	d := CanRead(p, r)
	if !d.Allowed {
		return d
	}
	switch p.Role {
	case "admin", "agronomist", "operator", "support":
		return Allow()
	default:
		return Deny("role_forbidden", fmt.Sprintf("role %s cannot mutate", p.Role))
	}
}
func ValidateWindow(start, end time.Time) error {
	if start.IsZero() || end.IsZero() {
		return fmt.Errorf("window must be set")
	}
	if !end.After(start) {
		return fmt.Errorf("window must end after start")
	}
	if end.Sub(start) > 14*24*time.Hour {
		return fmt.Errorf("window exceeds fourteen days")
	}
	return nil
}
func NormalizeFilter(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		v := strings.ToLower(strings.TrimSpace(part))
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
func IsTerminal(state string) bool {
	switch state {
	case "closed", "resolved", "cancelled", "fulfilled", "paid", "rejected", "failed":
		return true
	default:
		return false
	}
}
func CanTransition(from, to string, allowed map[string][]string) bool {
	for _, candidate := range allowed[from] {
		if candidate == to {
			return true
		}
	}
	return false
}
func RequireFresh(now, updated time.Time, ttl time.Duration) Decision {
	if updated.IsZero() {
		return Deny("missing_timestamp", "resource has no update timestamp")
	}
	if now.Sub(updated) > ttl {
		return Deny("stale", "resource is stale")
	}
	return Allow()
}
