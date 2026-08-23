package repository

import (
	"testing"
	"time"
)

func TestListArgs(t *testing.T) {
	q, a := ListArgs(QueryPage{Limit: 20, Offset: 40})
	if q == "" || len(a) != 2 || a[0] != 20 || a[1] != 40 {
		t.Fatalf("%s %v", q, a)
	}
}
func TestRetryDelay(t *testing.T) {
	if RetryDelay(0) != 250*time.Millisecond || RetryDelay(9) != RetryDelay(8) {
		t.Fatal("delay")
	}
}
