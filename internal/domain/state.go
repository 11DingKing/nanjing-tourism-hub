package domain

import "fmt"

func CanMoveIncident(from, to IncidentStatus) bool {
	allowed := map[IncidentStatus]map[IncidentStatus]bool{
		IncidentDraft:       {IncidentActive: true},
		IncidentActive:      {IncidentStabilizing: true},
		IncidentStabilizing: {IncidentActive: true, IncidentClosed: true},
		IncidentClosed:      {IncidentArchived: true},
	}
	return allowed[from][to]
}

func MoveIncident(current, target IncidentStatus) error {
	if current == target {
		return nil
	}
	if !CanMoveIncident(current, target) {
		return fmt.Errorf("%w: incident %s to %s", ErrInvalidState, current, target)
	}
	return nil
}

func CanMoveTask(from, to TaskStatus) bool {
	allowed := map[TaskStatus]map[TaskStatus]bool{
		TaskPending:   {TaskAccepted: true, TaskCanceled: true},
		TaskAccepted:  {TaskExecuting: true, TaskCanceled: true},
		TaskExecuting: {TaskCompleted: true, TaskCanceled: true},
	}
	return from == to || allowed[from][to]
}

func MoveTask(current, target TaskStatus) error {
	if !CanMoveTask(current, target) {
		return fmt.Errorf("%w: task %s to %s", ErrInvalidState, current, target)
	}
	return nil
}

func ResponseRank(level ResponseLevel) int {
	switch level {
	case ResponseIV:
		return 1
	case ResponseIII:
		return 2
	case ResponseII:
		return 3
	case ResponseI:
		return 4
	default:
		return 0
	}
}

func Escalates(from, to ResponseLevel) bool { return ResponseRank(to) > ResponseRank(from) }

func ValidRole(role Role) bool {
	switch role {
	case RoleDutyOfficer, RoleDispatcher, RoleReviewer, RoleFieldLead:
		return true
	default:
		return false
	}
}
