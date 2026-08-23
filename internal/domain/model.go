package domain

import "time"

type Role string

const (
	RoleDutyOfficer Role = "duty_officer"
	RoleDispatcher  Role = "dispatcher"
	RoleReviewer    Role = "reviewer"
	RoleFieldLead   Role = "field_lead"
)

type WarningLevel string

const (
	WarningBlue   WarningLevel = "blue"
	WarningYellow WarningLevel = "yellow"
	WarningOrange WarningLevel = "orange"
	WarningRed    WarningLevel = "red"
)

type IncidentStatus string

const (
	IncidentDraft       IncidentStatus = "draft"
	IncidentActive      IncidentStatus = "active"
	IncidentStabilizing IncidentStatus = "stabilizing"
	IncidentClosed      IncidentStatus = "closed"
	IncidentArchived    IncidentStatus = "archived"
)

type ResponseLevel string

const (
	ResponseIV  ResponseLevel = "IV"
	ResponseIII ResponseLevel = "III"
	ResponseII  ResponseLevel = "II"
	ResponseI   ResponseLevel = "I"
)

type TaskStatus string

const (
	TaskPending   TaskStatus = "pending"
	TaskAccepted  TaskStatus = "accepted"
	TaskExecuting TaskStatus = "executing"
	TaskCompleted TaskStatus = "completed"
	TaskCanceled  TaskStatus = "canceled"
)

type ResourceStatus string

const (
	ResourceAvailable   ResourceStatus = "available"
	ResourceReserved    ResourceStatus = "reserved"
	ResourceDeployed    ResourceStatus = "deployed"
	ResourceMaintenance ResourceStatus = "maintenance"
)

type User struct {
	ID, Username, PasswordHash string
	Role                       Role
	RegionID                   string
	Active                     bool
	Version                    int64
	CreatedAt, UpdatedAt       time.Time
}
type Session struct {
	ID, UserID, TokenHash string
	ExpiresAt             time.Time
	RevokedAt             *time.Time
	CreatedAt             time.Time
}
type Region struct {
	ID, Name, ParentID, TimeZone string
	Coastal                      bool
	RiskLevel                    int
	Version                      int64
	CreatedAt, UpdatedAt         time.Time
}
type Warning struct {
	ID, TyphoonName, Number                 string
	Level                                   WarningLevel
	IssuedAt, EffectiveFrom, EffectiveUntil time.Time
	RainfallMM                              int
	CenterArea, Movement                    string
	Status                                  string
	Version                                 int64
}
type Incident struct {
	ID, WarningID, Name  string
	Status               IncidentStatus
	CommanderID          string
	ActivatedAt          time.Time
	ClosedAt             *time.Time
	Version              int64
	CreatedAt, UpdatedAt time.Time
}
type Response struct {
	ID, IncidentID, RegionID string
	Level                    ResponseLevel
	Status                   string
	ActivatedBy              string
	ActivatedAt              time.Time
	Reason                   string
	Version                  int64
	UpdatedAt                time.Time
}
type Shelter struct {
	ID, RegionID, Name           string
	Capacity, Reserved, Occupied int
	Status                       ResourceStatus
	Version                      int64
	UpdatedAt                    time.Time
}
type Reservation struct {
	ID, ShelterID, IncidentID, RegionID string
	People                              int
	Status                              string
	IdempotencyKey                      string
	ExpiresAt                           time.Time
	Version                             int64
	CreatedAt, UpdatedAt                time.Time
}
type Team struct {
	ID, RegionID, Name, Specialty string
	Status                        ResourceStatus
	CurrentIncidentID             string
	LeaderID                      string
	Version                       int64
	UpdatedAt                     time.Time
}
type Dispatch struct {
	ID, TeamID, IncidentID, RegionID, TaskID string
	Status                                   TaskStatus
	RequestedBy                              string
	AcceptedAt, CompletedAt                  *time.Time
	Version                                  int64
	CreatedAt, UpdatedAt                     time.Time
}
type SupplyLot struct {
	ID, RegionID, Kind         string
	Quantity, Reserved, Issued int
	ExpiresAt                  time.Time
	Status                     ResourceStatus
	Version                    int64
	UpdatedAt                  time.Time
}
type Allocation struct {
	ID, LotID, IncidentID, RegionID, TaskID string
	Quantity                                int
	Status                                  string
	IdempotencyKey                          string
	Version                                 int64
	CreatedAt, UpdatedAt                    time.Time
}
type DistrictTask struct {
	ID, IncidentID, RegionID, Kind, Summary string
	Status                                  TaskStatus
	Priority                                int
	AssigneeID, CreatedBy                   string
	Deadline                                time.Time
	Version                                 int64
	CreatedAt, UpdatedAt                    time.Time
}
type Receipt struct {
	ID, TaskID, IncidentID, RegionID, ReporterID, Summary string
	Status                                                string
	EvidenceCount                                         int
	SubmittedAt                                           time.Time
	ReviewedAt                                            *time.Time
	Version                                               int64
}
type AuditEvent struct {
	ID, ActorID, RequestID, Action, Entity, EntityID, Result, Detail string
	CreatedAt                                                        time.Time
}
type OutboxEvent struct {
	ID, Topic, AggregateType, AggregateID string
	Payload                               []byte
	Status                                string
	Attempts                              int
	NextAttemptAt                         time.Time
	LastError                             string
	CreatedAt, UpdatedAt                  time.Time
}
type Job struct {
	ID, Kind, AggregateID string
	Payload               []byte
	Status                string
	Attempts, MaxAttempts int
	NextRunAt             time.Time
	LastError             string
	LockedUntil           *time.Time
	CreatedAt, UpdatedAt  time.Time
}
type IdempotencyRecord struct {
	Scope, Key, RequestHash string
	StatusCode              int
	Response                []byte
	ExpiresAt, CreatedAt    time.Time
}

type Page struct {
	Limit, Offset   int
	Sort, Direction string
}
type TaskFilter struct {
	IncidentID, RegionID string
	Status               TaskStatus
	Kind                 string
	DeadlineBefore       *time.Time
	Page                 Page
}
type TaskPage struct {
	Items []DistrictTask
	Total int
}
type OperationResult struct {
	ID, Status string
	Version    int64
}
type BatchItemResult struct{ InputID, ResourceID, Status, ErrorCode string }
type BatchResult struct {
	Items             []BatchItemResult
	Succeeded, Failed int
}
