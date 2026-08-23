package tourism
import "testing"
func TestTestPartnerListIsolation(t *testing.T){s:=&TaskState{Items:[]Agency{{City:"Kuala Lumpur"}}};x:=s.Run();x[0].City="altered";if s.Items[0].City!="Kuala Lumpur"{t.Fatal("alias")}}
