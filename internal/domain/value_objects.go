package domain

import (
	"fmt"
	"strings"
)

type RegionPath []string

func (p RegionPath) Contains(id string) bool {
	for _, item := range p {
		if item == id {
			return true
		}
	}
	return false
}
func ValidateRole(role string) error {
	switch role {
	case "admin", "agronomist", "operator", "auditor", "support":
		return nil
	default:
		return fmt.Errorf("unknown role %q", role)
	}
}
func NormalizeCrop(crop string) string { return strings.ToUpper(strings.TrimSpace(crop)) }
func IsHighRisk(level string) bool     { return level == "high" || level == "critical" }
