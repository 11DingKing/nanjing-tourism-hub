package tourism
import "testing"
func TestTestProductPaginationContract(t *testing.T){s:=&TaskState{Market:"malaysia"};s.Run("x");if s.Total!=1||len(s.Rows)!=1{t.Fatal("scope")}}
