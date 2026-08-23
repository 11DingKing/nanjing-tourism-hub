package clock

import (
	"testing"
	"time"
)

func TestClockOperations(t *testing.T) {
	base := time.Date(2026, 8, 21, 15, 4, 0, 0, time.FixedZone("CST", 8*3600))
	if DayStart(base).Hour() != 0 || DayEnd(base).Hour() != 23 {
		t.Fatal("day bounds")
	}
	if !InWindow(base, base.Add(-time.Hour), base.Add(time.Hour)) {
		t.Fatal("window")
	}
	if !Expired(base, base) {
		t.Fatal("deadline")
	}
	if AddBusinessDays(time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC), 1).Weekday() != time.Monday {
		t.Fatal("business day")
	}
}
