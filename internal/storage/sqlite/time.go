package sqlite

import (
	"database/sql"
	"time"
)

func stamp(value time.Time) string { return value.UTC().Format(time.RFC3339Nano) }
func parseStamp(value string) time.Time {
	parsed, _ := time.Parse(time.RFC3339Nano, value)
	return parsed
}
func nullableStamp(value *time.Time) any {
	if value == nil {
		return nil
	}
	return stamp(*value)
}
func scanTime(raw string, target *time.Time) { *target = parseStamp(raw) }
func scanOptional(raw sql.NullString) *time.Time {
	if !raw.Valid {
		return nil
	}
	value := parseStamp(raw.String)
	return &value
}
func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
