package forecast
import("testing";"time")
func TestTask014(t *testing.T){n:=time.Now();c:=NewCatalog();evidence:=[]Reading{{RegionID:"r",ObservedAt:n}};_ = c.Put(Warning{RegionID:"r",Code:"x",Level:"high",ValidFrom:n,ValidUntil:n.Add(time.Hour),Evidence:evidence});evidence[0].RainMM=9;z,_:=c.Get("r");if z.Evidence[0].RainMM==9{t.Fatal("alias")}}