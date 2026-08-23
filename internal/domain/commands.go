package domain

import "time"

type Actor struct {
	UserID              string
	Role                Role
	RegionID, RequestID string
}
type PublishWarningCommand struct {
	Warning Warning
	Actor   Actor
}
type ActivateIncidentCommand struct {
	IncidentID, WarningID, Name string
	Regions                     []string
	Level                       ResponseLevel
	Actor                       Actor
}
type EscalateResponseCommand struct {
	IncidentID, RegionID string
	FromVersion          int64
	Target               ResponseLevel
	Reason               string
	Actor                Actor
}
type ReserveShelterCommand struct {
	ReservationID, ShelterID, IncidentID, RegionID, IdempotencyKey string
	People                                                         int
	ExpiresAt                                                      time.Time
	Actor                                                          Actor
}
type ReleaseShelterCommand struct {
	ReservationID   string
	ExpectedVersion int64
	Actor           Actor
}
type DispatchTeamCommand struct {
	DispatchID, TeamID, IncidentID, RegionID, TaskID string
	TeamVersion                                      int64
	Actor                                            Actor
}
type HandoffTeamCommand struct {
	DispatchID, FromTeamID, ToTeamID string
	FromVersion, ToVersion           int64
	Actor                            Actor
}
type AllocateSupplyCommand struct {
	AllocationID, LotID, IncidentID, RegionID, TaskID, IdempotencyKey string
	Quantity                                                          int
	LotVersion                                                        int64
	Actor                                                             Actor
}
type CancelAllocationCommand struct {
	AllocationID      string
	AllocationVersion int64
	Actor             Actor
}
type CreateTaskCommand struct {
	Task  DistrictTask
	Actor Actor
}
type AdvanceTaskCommand struct {
	TaskID          string
	ExpectedVersion int64
	Target          TaskStatus
	Actor           Actor
}
type SubmitReceiptCommand struct {
	Receipt     Receipt
	TaskVersion int64
	Actor       Actor
}
type ReviewReceiptCommand struct {
	ReceiptID       string
	ExpectedVersion int64
	Accept          bool
	Note            string
	Actor           Actor
}
type CloseIncidentCommand struct {
	IncidentID      string
	ExpectedVersion int64
	Actor           Actor
}
type ArchiveIncidentCommand struct {
	IncidentID      string
	ExpectedVersion int64
	Actor           Actor
}
type BatchAssignItem struct {
	TaskID, TeamID           string
	TaskVersion, TeamVersion int64
}
type BatchAssignCommand struct {
	IncidentID string
	Items      []BatchAssignItem
	Actor      Actor
}
