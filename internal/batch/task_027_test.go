package batch
import("context";"testing")
func TestTask027(t *testing.T){c,cancel:=context.WithCancel(context.Background());cancel();p:=Processor[int]{Workers:1,Handle:func(context.Context,int)error{return nil}};r:=p.Run(c,[]int{1});if r[0].Err==nil{t.Fatal("cancel ignored")}}