package domain

import "time"

type User struct {
	ID, Name, Role, PasswordHash string
	Disabled                     bool
	CreatedAt                    time.Time
}
type Session struct {
	ID, UserID string
	ExpiresAt  time.Time
	RevokedAt  *time.Time
}
type Region struct {
	ID, Name, ParentID string
	FloodRisk          string
}
type Campaign struct {
	ID, RegionID, Name, Status string
	StartAt, EndAt             time.Time
	Version                    int
}
type Plot struct {
	ID, RegionID, OwnerID string
	Acres                 float64
	Crop                  string
	Status                string
	Version               int
}
type Observation struct {
	ID, PlotID, ObserverID, Kind, Severity, Notes string
	ObservedAt                                    time.Time
}
type Alert struct {
	ID, RegionID, CampaignID, Level, Status, Message string
	CreatedAt, AcknowledgedAt                        *time.Time
}
type FieldTask struct {
	ID, PlotID, AssigneeID, Kind, Status, LeaseID string
	DueAt                                         time.Time
	Version                                       int
}
type DisasterReport struct {
	ID, RegionID, ReporterID, Kind, Status, Description string
	CreatedAt                                           time.Time
}
type RecoveryCase struct {
	ID, ReportID, RegionID, Status, OwnerID string
	EstimatedLoss                           float64
	Version                                 int
}
type DryingSite struct {
	ID, RegionID, Name string
	CapacityTons       float64
	Active             bool
	Version            int
}
type DryingReservation struct {
	ID, SiteID, PlotID, Status string
	Tons                       float64
	StartsAt, EndsAt           time.Time
	Version                    int
}
type AidCase struct {
	ID, RegionID, FarmerID, Status, Reason string
	AmountCents                            int64
	Version                                int
}
type FarmProject struct {
	ID, RegionID, Name, Status string
	BudgetCents                int64
	Version                    int
}
type AuditEvent struct {
	ID, ActorID, Action, ObjectType, ObjectID, Result, RequestID string
	CreatedAt                                                    time.Time
}
type OutboxEvent struct {
	ID, Topic, AggregateID, Payload, Status string
	Attempts                                int
	AvailableAt                             time.Time
}
