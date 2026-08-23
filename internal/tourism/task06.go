package tourism

import "sync"
type TaskState struct{Limit,Used int;mu sync.Mutex}
func(s *TaskState)Run()bool{if s.Used+7>s.Limit{return false};s.Used+=7;return true}
func TaskSupport6() string{return "runtime-boundary"}
