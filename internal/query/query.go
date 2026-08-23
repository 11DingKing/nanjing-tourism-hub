package query

import (
	"fmt"
	"strconv"
	"strings"
)

type Page struct {
	Number int
	Size   int
}
type Sort struct {
	Field      string
	Descending bool
}
type Filter struct {
	RegionID string
	Statuses []string
	Search   string
}
type SQL struct {
	Text string
	Args []any
}

func ParsePage(rawNumber, rawSize string) (Page, error) {
	n, s := 1, 50
	var err error
	if rawNumber != "" {
		n, err = strconv.Atoi(rawNumber)
		if err != nil || n < 1 {
			return Page{}, fmt.Errorf("invalid page")
		}
	}
	if rawSize != "" {
		s, err = strconv.Atoi(rawSize)
		if err != nil || s < 1 || s > 200 {
			return Page{}, fmt.Errorf("invalid page size")
		}
	}
	return Page{n, s}, nil
}
func (p Page) Offset() int { return (p.Number - 1) * p.Size }
func ParseSort(raw string, allowed map[string]bool) (Sort, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Sort{Field: "created_at"}, nil
	}
	desc := strings.HasPrefix(raw, "-")
	field := strings.TrimPrefix(raw, "-")
	if !allowed[field] {
		return Sort{}, fmt.Errorf("sort field %s is not allowed", field)
	}
	return Sort{Field: field, Descending: desc}, nil
}
func BuildWhere(f Filter) (string, []any) {
	parts := []string{"1=1"}
	args := []any{}
	if f.RegionID != "" {
		parts = append(parts, "region_id = ?")
		args = append(args, f.RegionID)
	}
	if len(f.Statuses) > 0 {
		marks := make([]string, len(f.Statuses))
		for i, status := range f.Statuses {
			marks[i] = "?"
			args = append(args, status)
		}
		parts = append(parts, "status IN ("+strings.Join(marks, ",")+")")
	}
	if f.Search != "" {
		parts = append(parts, "name LIKE ?")
		args = append(args, "%"+f.Search+"%")
	}
	return strings.Join(parts, " AND "), args
}
func BuildList(table string, f Filter, p Page, s Sort) (SQL, error) {
	allowed := map[string]bool{"created_at": true, "name": true, "status": true}
	sort, err := ParseSort(s.Field, allowed)
	if err != nil {
		return SQL{}, err
	}
	where, args := BuildWhere(f)
	direction := "ASC"
	if sort.Descending {
		direction = "DESC"
	}
	text := fmt.Sprintf("SELECT * FROM %s WHERE %s ORDER BY %s %s LIMIT ? OFFSET ?", table, where, sort.Field, direction)
	args = append(args, p.Size, p.Offset())
	return SQL{Text: text, Args: args}, nil
}
func CountSQL(table string, f Filter) (SQL, string, error) {
	where, args := BuildWhere(f)
	if !strings.HasPrefix(table, "safe_") {
		return SQL{}, "", fmt.Errorf("table is not allowlisted")
	}
	return SQL{Text: "SELECT * FROM " + table + " WHERE " + where, Args: args}, "SELECT COUNT(*) FROM " + table + " WHERE " + where, nil
}
