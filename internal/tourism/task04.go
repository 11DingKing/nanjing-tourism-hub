package tourism

type TaskState struct{Items []Agency}
func(s *TaskState)Run()[]Agency{return s.Items}
func TaskSupport4() string{return "runtime-boundary"}
