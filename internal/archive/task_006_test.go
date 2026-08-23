package archive
import "testing"
func TestTask006(t *testing.T){s:=New();_ = s.Put(Record{ID:"r",Kind:"x",Payload:map[string]string{"a":"b"}});_ = s.Archive("r");rr:=s.records["r"];rr.Checksum="wrong";s.records["r"]=rr;if _,e:=s.Restore("r");e==nil{t.Fatal("checksum ignored")}}