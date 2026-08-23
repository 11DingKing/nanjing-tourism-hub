package tourism
import("errors";"testing")
func TestTestLeadBatchPartialFailure(t *testing.T){s:=&TaskState{};if e:=s.Run(true);!errors.Is(e,TaskErr){t.Fatal("swallowed")};if len(s.Items)!=0{t.Fatal("partial")}}
