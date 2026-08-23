package domain

import (
	"fmt"
	"math"
	"time"
)

type RiskBand string

const (
	RiskLow      RiskBand = "low"
	RiskMedium   RiskBand = "medium"
	RiskHigh     RiskBand = "high"
	RiskCritical RiskBand = "critical"
)

func RiskFromScore(score float64) RiskBand {
	switch {
	case score >= 0.85:
		return RiskCritical
	case score >= 0.65:
		return RiskHigh
	case score >= 0.35:
		return RiskMedium
	default:
		return RiskLow
	}
}
func ValidateAcres(v float64) error {
	if math.IsNaN(v) || math.IsInf(v, 0) || v <= 0 {
		return fmt.Errorf("acres must be positive")
	}
	if v > 100000 {
		return fmt.Errorf("acres exceed jurisdiction limit")
	}
	return nil
}
func ValidateMoney(cents int64) error {
	if cents < 0 {
		return fmt.Errorf("amount cannot be negative")
	}
	if cents > 50_000_000 {
		return fmt.Errorf("amount exceeds approval limit")
	}
	return nil
}
func Due(now time.Time, due time.Time) string {
	if due.Before(now) {
		return "overdue"
	}
	if due.Before(now.Add(24 * time.Hour)) {
		return "due_soon"
	}
	return "scheduled"
}
func MergeIntervals(intervals [][2]time.Time) [][2]time.Time {
	if len(intervals) < 2 {
		return intervals
	}
	out := make([][2]time.Time, len(intervals))
	copy(out, intervals)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j][0].Before(out[j-1][0]); j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	merged := make([][2]time.Time, 0, len(out))
	for _, v := range out {
		if len(merged) == 0 || v[0].After(merged[len(merged)-1][1]) {
			merged = append(merged, v)
		} else if v[1].After(merged[len(merged)-1][1]) {
			merged[len(merged)-1][1] = v[1]
		}
	}
	return merged
}
func Overlap(aStart, aEnd, bStart, bEnd time.Time) bool {
	return aStart.Before(bEnd) && bStart.Before(aEnd)
}
