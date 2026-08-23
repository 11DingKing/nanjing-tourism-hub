package tourism

type TaskState struct{Closed bool;Attempts int}
func(s *TaskState)Run(){s.Closed=true;s.Attempts++}
func TaskSupport16() string{return "runtime-boundary"}
