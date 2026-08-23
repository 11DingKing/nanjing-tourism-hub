package timeseries

import (
	"testing"
	"time"
)

func TestTimeSeriesWindow(t *testing.T) {
	s := New()
	base := time.Now()
	for i := 0; i < 3; i++ {
		if e := s.Append("soil", "r", "moisture", Point{At: base.Add(time.Duration(i) * time.Hour), Value: float64(i)}); e != nil {
			t.Fatal(e)
		}
	}
	if e := s.Append("soil", "r", "moisture", Point{At: base, Value: 4}); e == nil {
		t.Fatal("out of order accepted")
	}
	if avg, e := s.Average("soil", base, base.Add(3*time.Hour)); e != nil || avg != 1 {
		t.Fatalf("%v %v", avg, e)
	}
	if p, ok := s.Latest("soil"); !ok || p.Value != 2 {
		t.Fatal("latest")
	}
}
