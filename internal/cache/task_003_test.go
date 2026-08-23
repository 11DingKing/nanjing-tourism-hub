package cache
import("testing";"time")
func TestTask003(t *testing.T){c:=New();_ = c.Set("k","old",-time.Second,1);if !c.CompareAndSet("k","new",time.Hour,2,3){t.Fatal("expired entry blocks replacement")}}