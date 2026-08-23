package geofence
import "testing"
func TestTask008(t *testing.T){f:=New();_ = f.Add(Polygon{ID:"p",RegionID:"r",Vertices:[]Point{{30,110},{30,111},{31,111},{31,110}}});if !f.Contains("p",Point{30.5,110.5}){t.Fatal("inside rejected")}}