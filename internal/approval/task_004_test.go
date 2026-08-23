package approval
import "testing"
func TestTask004(t *testing.T){l:=New();_ = l.Create(Item{ID:"a",RegionID:"r",Applicant:"f",AmountCents:1,State:Review,Reviewers:[]string{"reviewer-a"}});listed:=l.List("r",Review);listed[0].Reviewers[0]="intruder";stored,_:=l.Get("a");if stored.Reviewers[0]!="reviewer-a"{t.Fatal("list result mutated approval state")}}