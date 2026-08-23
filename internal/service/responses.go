package service

import (
	"context"
	"fmt"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
	"github.com/11DingKing/nanjing-tourism-hub/internal/repository"
)

func (c *Coordinator) EscalateResponse(ctx context.Context, cmd domain.EscalateResponseCommand) (domain.Response, error) {
	if err := domain.RequireRole(cmd.Actor, domain.RoleDutyOfficer, domain.RoleDispatcher); err != nil {
		return domain.Response{}, err
	}
	if err := requireRegion(cmd.Actor, cmd.RegionID); err != nil {
		return domain.Response{}, err
	}
	var changed domain.Response
	err := c.repo.WithinTx(ctx, func(tx repository.Tx) error {
		incident, err := tx.GetIncident(ctx, cmd.IncidentID)
		if err != nil {
			return err
		}
		if incident.Status != domain.IncidentActive {
			return domain.Wrap(domain.ErrInvalidState, "escalate", "incident", incident.ID, "incident is not active", nil)
		}
		current, err := tx.GetResponse(ctx, cmd.IncidentID, cmd.RegionID)
		if err != nil {
			return err
		}
		if !domain.Escalates(current.Level, cmd.Target) {
			return domain.Wrap(domain.ErrInvalidState, "escalate", "response", current.ID, "target must increase response level", nil)
		}
		if cmd.FromVersion != current.Version {
			return domain.VersionConflict("response", current.ID, cmd.FromVersion, current.Version)
		}
		changed = current
		changed.Level = cmd.Target
		changed.Reason = cmd.Reason
		changed.ActivatedBy = cmd.Actor.UserID
		changed.ActivatedAt = c.now().UTC()
		changed.UpdatedAt = changed.ActivatedAt
		changed.Version++
		if err := tx.UpdateResponse(ctx, changed, current.Version); err != nil {
			return err
		}
		if err := c.outbox(ctx, tx, "response.escalated", "response", changed.ID, changed); err != nil {
			return err
		}
		return c.audit(ctx, tx, cmd.Actor, "response.escalate", "response", changed.ID, "success", fmt.Sprintf("%s to %s", current.Level, changed.Level))
	})
	if err != nil {
		return domain.Response{}, fmt.Errorf("escalate response: %w", err)
	}
	return changed, nil
}
