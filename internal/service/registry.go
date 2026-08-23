package service

import (
	"database/sql"
	"github.com/11DingKing/autumn-grain-resilience/internal/config"
	"github.com/11DingKing/autumn-grain-resilience/internal/repository"
	"log/slog"
)

type Registry struct {
	Auth       *AuthService
	Monitoring *MonitoringService
	FieldTasks *FieldTaskService
	Disasters  *DisasterService
	Drying     *DryingService
	Aid        *AidService
	Audit      *AuditService
	Outbox     *repository.OutboxRepo
}

func NewRegistry(db *sql.DB, logger *slog.Logger, cfg config.Config) *Registry {
	audit := &AuditService{Repo: repository.AuditRepo{DB: db}, Logger: logger}
	out := &repository.OutboxRepo{DB: db}
	return &Registry{Auth: &AuthService{Users: repository.UserRepo{DB: db}, TTL: cfg.SessionTTL}, Monitoring: &MonitoringService{DB: db, Audit: audit, Outbox: out}, FieldTasks: &FieldTaskService{Repo: repository.TaskRepo{DB: db}, Plots: repository.PlotRepo{DB: db}, Audit: audit}, Disasters: &DisasterService{DB: db, Repo: repository.DisasterRepo{DB: db}, Audit: audit}, Drying: &DryingService{DB: db, Repo: repository.DryingRepo{DB: db}, Plots: repository.PlotRepo{DB: db}, Audit: audit}, Aid: &AidService{Repo: repository.AidRepo{DB: db}, Audit: audit}, Audit: audit, Outbox: out}
}
