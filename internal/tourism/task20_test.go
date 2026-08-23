package tourism
import("context";"testing")
func TestTestMalaysiaLocalCutoff(t *testing.T){s:=&TaskState{};x,y:=context.WithCancel(context.Background());y();if e:=s.Run(x);e==nil{t.Fatal("want cancel")};if len(s.Items)!=0{t.Fatal("partial")}}
