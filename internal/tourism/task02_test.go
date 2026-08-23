package tourism
import("sync";"testing")
func TestTestConcurrentGroupQuota(t *testing.T){s:=&TaskState{Limit:10};var w sync.WaitGroup;ch:=make(chan bool,2);for i:=0;i<2;i++{w.Add(1);go func(){defer w.Done();ch<-s.Run()}()};w.Wait();close(ch);n:=0;for x:=range ch{if x{n++}};if n!=1||s.Used>10{t.Fatal("oversold")}}
