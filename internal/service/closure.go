package service

import (
	"context"
	"fmt"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
	"github.com/11DingKing/nanjing-tourism-hub/internal/repository"
)

func (c *Coordinator) CloseIncident(ctx context.Context, cmd domain.CloseIncidentCommand) (domain.Incident, error) {
	if err := domain.RequireRole(cmd.Actor, domain.RoleDutyOfficer); err != nil {
		return domain.Incident{}, err
	}
	var changed domain.Incident
	err := c.repo.WithinTx(ctx, func(tx repository.Tx) error {
		incident, err := tx.GetIncident(ctx, cmd.IncidentID)
		if err != nil {
			return err
		}
		if incident.Status == domain.IncidentActive {
			incident.Status = domain.IncidentStabilizing
		}
		if err := domain.MoveIncident(incident.Status, domain.IncidentClosed); err != nil {
			return err
		}
		open, err := tx.CountOpenTasks(ctx, incident.ID)
		if err != nil {
			return err
		}
		pending, err := tx.CountPendingReceipts(ctx, incident.ID)
		if err != nil {
			return err
		}
		if open > 0 || pending > 0 {
			return domain.Wrap(domain.ErrConflict, "close", "incident", incident.ID, fmt.Sprintf("open_tasks=%d pending_receipts=%d", open, pending), nil)
		}
		now := c.now().UTC()
		incident.Status = domain.IncidentClosed
		incident.ClosedAt = &now
		incident.UpdatedAt = now
		incident.Version++
		if err := tx.UpdateIncident(ctx, incident, cmd.ExpectedVersion); err != nil {
			return err
		}
		if err := c.outbox(ctx, tx, "incident.closed", "incident", incident.ID, incident); err != nil {
			return err
		}
		changed = incident
		return c.audit(ctx, tx, cmd.Actor, "incident.close", "incident", incident.ID, "success", "")
	})
	if err != nil {
		return domain.Incident{}, fmt.Errorf("close incident: %w", err)
	}
	return changed, nil
}
func (c *Coordinator) ArchiveIncident(ctx context.Context, cmd domain.ArchiveIncidentCommand) (domain.Incident, error) {
	if err := domain.RequireRole(cmd.Actor, domain.RoleReviewer); err != nil {
		return domain.Incident{}, err
	}
	var changed domain.Incident
	err := c.repo.WithinTx(ctx, func(tx repository.Tx) error {
		incident, err := tx.GetIncident(ctx, cmd.IncidentID)
		if err != nil {
			return err
		}
		if err := domain.MoveIncident(incident.Status, domain.IncidentArchived); err != nil {
			return err
		}
		incident.Status = domain.IncidentArchived
		incident.UpdatedAt = c.now().UTC()
		incident.Version++
		if err := tx.UpdateIncident(ctx, incident, cmd.ExpectedVersion); err != nil {
			return err
		}
		if err := c.job(ctx, tx, "incident.retention", incident.ID, map[string]string{"incident_id": incident.ID}, c.now().Add(24*60*60*1000000000)); err != nil {
			return err
		}
		changed = incident
		return c.audit(ctx, tx, cmd.Actor, "incident.archive", "incident", incident.ID, "success", "")
	})
	if err != nil {
		return domain.Incident{}, fmt.Errorf("archive incident: %w", err)
	}
	return changed, nil
}
