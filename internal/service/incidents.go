package service

import (
	"context"
	"fmt"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
	"github.com/11DingKing/nanjing-tourism-hub/internal/repository"
	"time"
)

func (c *Coordinator) ActivateIncident(ctx context.Context, cmd domain.ActivateIncidentCommand) (domain.Incident, error) {
	if err := domain.RequireRole(cmd.Actor, domain.RoleDutyOfficer); err != nil {
		return domain.Incident{}, err
	}
	if len(cmd.Regions) == 0 {
		return domain.Incident{}, fmt.Errorf("%w: at least one region is required", domain.ErrInvalidState)
	}
	now := c.now().UTC()
	incident := domain.Incident{ID: cmd.IncidentID, WarningID: cmd.WarningID, Name: cmd.Name, Status: domain.IncidentActive, CommanderID: cmd.Actor.UserID, ActivatedAt: now, Version: 1, CreatedAt: now, UpdatedAt: now}
	if incident.ID == "" {
		incident.ID = newID("incident")
	}
	err := c.repo.WithinTx(ctx, func(tx repository.Tx) error {
		warning, err := tx.GetWarning(ctx, cmd.WarningID)
		if err != nil {
			return err
		}
		if warning.Status != "published" || !warning.EffectiveUntil.After(now) {
			return domain.Wrap(domain.ErrInvalidState, "activate", "incident", incident.ID, "warning is not active", nil)
		}
		if err := tx.InsertIncident(ctx, incident); err != nil {
			return err
		}
		seen := map[string]struct{}{}
		for _, regionID := range cmd.Regions {
			if _, ok := seen[regionID]; ok {
				return domain.Wrap(domain.ErrConflict, "activate", "region", regionID, "duplicate region", nil)
			}
			seen[regionID] = struct{}{}
			response := domain.Response{ID: newID("response"), IncidentID: incident.ID, RegionID: regionID, Level: cmd.Level, Status: "active", ActivatedBy: cmd.Actor.UserID, ActivatedAt: now, Reason: "warning activation", Version: 1, UpdatedAt: now}
			if err := tx.InsertResponse(ctx, response); err != nil {
				return err
			}
		}
		if err := c.outbox(ctx, tx, "incident.activated", "incident", incident.ID, incident); err != nil {
			return err
		}
		if err := c.job(ctx, tx, "response.check", incident.ID, map[string]any{"incident_id": incident.ID}, now.Add(15*time.Minute)); err != nil {
			return err
		}
		return c.audit(ctx, tx, cmd.Actor, "incident.activate", "incident", incident.ID, "success", cmd.Name)
	})
	if err != nil {
		return domain.Incident{}, fmt.Errorf("activate incident: %w", err)
	}
	return incident, nil
}
