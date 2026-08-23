package repository

import (
	"context"
	"time"

	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
)

type Repository interface {
	WithinTx(context.Context, func(Tx) error) error
	Ping(context.Context) error
	Migrate(context.Context) error
	Close() error
	Readiness(context.Context) error
	FindSessionByHash(context.Context, string) (domain.Session, domain.User, error)
	FindUserByUsername(context.Context, string) (domain.User, error)
	RevokeSession(context.Context, string, time.Time) error
	ListTasks(context.Context, domain.TaskFilter) (domain.TaskPage, error)
	ClaimDueJobs(context.Context, time.Time, int, time.Duration) ([]domain.Job, error)
	ClaimOutbox(context.Context, time.Time, int) ([]domain.OutboxEvent, error)
	CompleteJob(context.Context, string, time.Time) error
	FailJob(context.Context, string, string, time.Time) error
	CompleteOutbox(context.Context, string, time.Time) error
	FailOutbox(context.Context, string, string, time.Time) error
	SnapshotIncident(context.Context, string) (IncidentSnapshot, error)
}

type Tx interface {
	InsertRegion(context.Context, domain.Region) error
	CreateUser(context.Context, domain.User) error
	CreateSession(context.Context, domain.Session) error
	RevokeSessionsForUser(context.Context, string, time.Time) error
	InsertWarning(context.Context, domain.Warning) error
	GetWarning(context.Context, string) (domain.Warning, error)
	UpdateWarning(context.Context, domain.Warning, int64) error
	InsertIncident(context.Context, domain.Incident) error
	GetIncident(context.Context, string) (domain.Incident, error)
	UpdateIncident(context.Context, domain.Incident, int64) error
	InsertResponse(context.Context, domain.Response) error
	GetResponse(context.Context, string, string) (domain.Response, error)
	UpdateResponse(context.Context, domain.Response, int64) error
	GetShelter(context.Context, string) (domain.Shelter, error)
	UpdateShelter(context.Context, domain.Shelter, int64) error
	InsertShelter(context.Context, domain.Shelter) error
	InsertReservation(context.Context, domain.Reservation) error
	GetReservation(context.Context, string) (domain.Reservation, error)
	UpdateReservation(context.Context, domain.Reservation, int64) error
	GetTeam(context.Context, string) (domain.Team, error)
	UpdateTeam(context.Context, domain.Team, int64) error
	InsertTeam(context.Context, domain.Team) error
	InsertDispatch(context.Context, domain.Dispatch) error
	GetDispatch(context.Context, string) (domain.Dispatch, error)
	UpdateDispatch(context.Context, domain.Dispatch, int64) error
	GetSupplyLot(context.Context, string) (domain.SupplyLot, error)
	UpdateSupplyLot(context.Context, domain.SupplyLot, int64) error
	InsertSupplyLot(context.Context, domain.SupplyLot) error
	InsertAllocation(context.Context, domain.Allocation) error
	GetAllocation(context.Context, string) (domain.Allocation, error)
	UpdateAllocation(context.Context, domain.Allocation, int64) error
	InsertTask(context.Context, domain.DistrictTask) error
	GetTask(context.Context, string) (domain.DistrictTask, error)
	UpdateTask(context.Context, domain.DistrictTask, int64) error
	InsertReceipt(context.Context, domain.Receipt) error
	GetReceipt(context.Context, string) (domain.Receipt, error)
	UpdateReceipt(context.Context, domain.Receipt, int64) error
	CountOpenTasks(context.Context, string) (int, error)
	CountPendingReceipts(context.Context, string) (int, error)
	InsertAudit(context.Context, domain.AuditEvent) error
	InsertOutbox(context.Context, domain.OutboxEvent) error
	InsertJob(context.Context, domain.Job) error
	GetIdempotency(context.Context, string, string) (domain.IdempotencyRecord, error)
	PutIdempotency(context.Context, domain.IdempotencyRecord) error
}

type IncidentSnapshot struct {
	Incident    domain.Incident
	Responses   []domain.Response
	Tasks       []domain.DistrictTask
	Receipts    []domain.Receipt
	Dispatches  []domain.Dispatch
	Allocations []domain.Allocation
}
