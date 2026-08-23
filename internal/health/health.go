package health

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Status struct {
	Healthy    bool
	Dependency string
	CheckedAt  time.Time
	Latency    time.Duration
	Message    string
}

func Check(ctx context.Context, db *sql.DB) Status {
	started := time.Now()
	s := Status{CheckedAt: started}
	if db == nil {
		s.Message = "database handle missing"
		return s
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	var value int
	if e := db.QueryRowContext(ctx, "SELECT 1").Scan(&value); e != nil {
		s.Message = e.Error()
		s.Latency = time.Since(started)
		return s
	}
	s.Healthy = value == 1
	s.Dependency = "database"
	s.Message = "ready"
	s.Latency = time.Since(started)
	return s
}
func Validate(s Status) error {
	if !s.Healthy {
		return errors.New("dependency unhealthy")
	}
	if s.Dependency == "" || s.CheckedAt.IsZero() {
		return errors.New("health evidence incomplete")
	}
	return nil
}
func Ready(ctx context.Context, db *sql.DB) error { return Validate(Check(ctx, db)) }
func DependencySummary(statuses []Status) Status {
	out := Status{Healthy: true, CheckedAt: time.Now().UTC(), Dependency: "composite", Message: "ready"}
	for _, s := range statuses {
		if !s.Healthy {
			out.Healthy = false
			out.Message = "one or more dependencies unavailable"
		}
		if s.Latency > out.Latency {
			out.Latency = s.Latency
		}
	}
	return out
}
func Probe(ctx context.Context, db *sql.DB, attempts int) Status {
	if attempts < 1 {
		attempts = 1
	}
	var last Status
	for i := 0; i < attempts; i++ {
		last = Check(ctx, db)
		if last.Healthy {
			return last
		}
		select {
		case <-ctx.Done():
			return last
		case <-time.After(time.Millisecond):
		}
	}
	return last
}
func IsDegraded(s Status, threshold time.Duration) bool { return s.Healthy && s.Latency > threshold }
func Merge(statuses ...Status) Status                   { return DependencySummary(statuses) }
func Deadline(ctx context.Context) (time.Time, bool)    { d, ok := ctx.Deadline(); return d, ok }
func Healthy(statuses []Status) bool {
	if len(statuses) == 0 {
		return false
	}
	for _, s := range statuses {
		if !s.Healthy {
			return false
		}
	}
	return true
}
func Message(s Status) string {
	if s.Healthy {
		return s.Dependency + " ready"
	}
	if s.Message != "" {
		return s.Message
	}
	return "dependency unavailable"
}
