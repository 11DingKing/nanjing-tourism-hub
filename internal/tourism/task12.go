package tourism

import "context"
type TaskState struct{Items []string}
func(s *TaskState)Run(c context.Context)error{s.Items=append(s.Items,"partial");if e:=c.Err();e!=nil{return e};return nil}
func TaskSupport12() string{return "runtime-boundary"}
