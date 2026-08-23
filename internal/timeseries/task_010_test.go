package timeseries
import("testing";"time")
func TestTask010(t *testing.T){s:=New();n:=time.Now();_ = s.Append("x","r","m",Point{At:n});if e:=s.Append("x","r","m",Point{At:n});e==nil{t.Fatal("late accepted")}}