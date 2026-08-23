package recovery
import("context";"testing")
func TestTask001(t *testing.T){p:=NewPlan("p","r","d",[]Step{{ID:"x",State:Assigned}});ctx,cancel:=context.WithCancel(context.Background());cancel();if e:=Run(ctx,p,"x");e==nil{t.Fatal("cancelled recovery executed")};if p.Snapshot()[0].State!=Assigned{t.Fatal("state advanced after cancellation")}}