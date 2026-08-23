package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/11DingKing/autumn-grain-resilience/internal/domain"
	"github.com/11DingKing/autumn-grain-resilience/internal/repository"
	"time"
)

type MonitoringService struct {
	DB     *sql.DB
	Audit  *AuditService
	Outbox *repository.OutboxRepo
}

func (s *MonitoringService) Record(ctx context.Context, obs domain.Observation, regionID, request string) (domain.Alert, error) {
	tx, e := s.DB.BeginTx(ctx, nil)
	if e != nil {
		return domain.Alert{}, e
	}
	defer tx.Rollback()
	if _, e = tx.ExecContext(ctx, "INSERT INTO observations(id,plot_id,observer_id,kind,severity,notes,observed_at) VALUES(?,?,?,?,?,?,?)", obs.ID, obs.PlotID, obs.ObserverID, obs.Kind, obs.Severity, obs.Notes, obs.ObservedAt.UTC().Format(time.RFC3339Nano)); e != nil {
		return domain.Alert{}, e
	}
	alert := domain.Alert{ID: id("alt_"), RegionID: regionID, Level: obs.Severity, Status: "open", Message: "crop stress requires review", CreatedAt: &obs.ObservedAt}
	if domain.IsHighRisk(obs.Severity) {
		if _, e = tx.ExecContext(ctx, "INSERT INTO alerts(id,region_id,level,status,message,created_at) VALUES(?,?,?,?,?,?)", alert.ID, alert.RegionID, alert.Level, alert.Status, alert.Message, obs.ObservedAt.UTC().Format(time.RFC3339Nano)); e != nil {
			return domain.Alert{}, e
		}
		payload, _ := json.Marshal(alert)
		if _, e = tx.ExecContext(ctx, "INSERT INTO outbox_events(id,topic,aggregate_id,payload,status,attempts,available_at) VALUES(?,?,?,?,?,?,?)", id("evt_"), "alert.created", alert.ID, string(payload), "pending", 0, time.Now().UTC().Format(time.RFC3339Nano)); e != nil {
			return domain.Alert{}, e
		}
	}
	if e = s.Audit.Repo.AppendTx(ctx, tx, domain.AuditEvent{ID: id("aud_"), ActorID: obs.ObserverID, Action: "record_observation", ObjectType: "observation", ObjectID: obs.ID, Result: "success", RequestID: request, CreatedAt: time.Now().UTC()}); e != nil {
		return domain.Alert{}, e
	}
	if e = tx.Commit(); e != nil {
		return domain.Alert{}, e
	}
	return alert, nil
}
