package tourism

import "time"
type TaskState struct{Market string;Rows []string;Total int;Cutoff time.Time}
func(s *TaskState)Run(v string){s.Rows=[]string{"malaysia","singapore"};s.Total=2}
func TaskSupport17() string{return "runtime-boundary"}
