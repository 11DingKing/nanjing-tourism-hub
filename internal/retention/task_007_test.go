package retention
import("testing";"time")
func TestTask007(t *testing.T){m:=New();_ = m.Rule(Rule{Kind:"evidence",Keep:time.Hour});_ = m.Add(Object{ID:"x",Kind:"evidence",CreatedAt:time.Now().Add(-2*time.Hour),Hold:true});if got:=m.Eligible(time.Now());len(got)!=0{t.Fatal("held evidence scheduled for deletion",got)}}