package forecast
import("testing";"time")
func TestTask022(t *testing.T){n:=time.Now();r:=Aggregate([]Reading{{RegionID:"r",ObservedAt:n,Temperature:20},{RegionID:"r",ObservedAt:n.Add(time.Minute),Temperature:22}});if r.Temperature!=21{t.Fatal(r.Temperature)}}