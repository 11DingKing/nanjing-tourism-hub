package clock

import "time"

type Clock interface{ Now() time.Time }
type Real struct{}

func (Real) Now() time.Time { return time.Now().UTC() }

type Fixed struct{ Value time.Time }

func (f Fixed) Now() time.Time { return f.Value }
func DayStart(t time.Time) time.Time {
	u := t.UTC()
	return time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
}
func DayEnd(t time.Time) time.Time            { return DayStart(t).Add(24 * time.Hour).Add(-time.Nanosecond) }
func InWindow(now, start, end time.Time) bool { return !now.Before(start) && now.Before(end) }
func Expired(now, deadline time.Time) bool    { return !now.Before(deadline) }
func Clamp(t, min, max time.Time) time.Time {
	if t.Before(min) {
		return min
	}
	if t.After(max) {
		return max
	}
	return t
}
func AddBusinessDays(t time.Time, n int) time.Time {
	step := 1
	if n < 0 {
		step = -1
		n = -n
	}
	for n > 0 {
		t = t.Add(time.Duration(step) * 24 * time.Hour)
		if t.Weekday() != time.Saturday && t.Weekday() != time.Sunday {
			n--
		}
	}
	return t
}
