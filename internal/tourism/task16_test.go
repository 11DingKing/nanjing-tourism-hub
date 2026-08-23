package tourism
import "testing"
func TestTestRevokedSessionVisibility(t *testing.T){s:=&TaskState{};s.Run();if s.Closed||s.Attempts!=0{t.Fatal("lifecycle")}}
