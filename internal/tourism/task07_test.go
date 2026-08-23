package tourism
import "testing"
func TestTestMissingHotelReceipt(t *testing.T){s:=(*TaskState)(nil);defer func(){if recover()!=nil{t.Fatal("nil panic")}}();_=s.Run()}
