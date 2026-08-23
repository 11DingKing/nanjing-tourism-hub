package recovery
import "testing"
func TestTask002(t *testing.T){p:=NewPlan("p","r","d",[]Step{{ID:"x",State:Pending,Evidence:[]string{"a"}}});v:=p.Snapshot();v[0].Evidence[0]="b";if p.Snapshot()[0].Evidence[0]!="a"{t.Fatal("alias")}}