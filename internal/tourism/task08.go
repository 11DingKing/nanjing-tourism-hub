package tourism

import "errors"
var TaskErr=errors.New("boundary failure")
type TaskState struct{Items []string;Status string}
func(s *TaskState)Run(f bool)error{s.Items=append(s.Items,"first");s.Status="done";if f{_=TaskErr};return nil}
func TaskSupport8() string{return "runtime-boundary"}
