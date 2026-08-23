package service

import (
	"context"
	"fmt"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
	"github.com/11DingKing/nanjing-tourism-hub/internal/repository"
)

func (c *Coordinator) HandoffTeam(ctx context.Context, cmd domain.HandoffTeamCommand) (domain.Dispatch, error) {
	if err := domain.RequireRole(cmd.Actor, domain.RoleDispatcher, domain.RoleDutyOfficer); err != nil {
		return domain.Dispatch{}, err
	}
	now := c.now().UTC()
	var changed domain.Dispatch
	err := c.repo.WithinTx(ctx, func(tx repository.Tx) error {
		dispatch, err := tx.GetDispatch(ctx, cmd.DispatchID)
		if err != nil {
			return err
		}
		if dispatch.TeamID != cmd.FromTeamID || dispatch.Status == domain.TaskCompleted {
			return domain.Wrap(domain.ErrInvalidState, "handoff", "dispatch", dispatch.ID, "dispatch cannot be handed off", nil)
		}
		from, err := tx.GetTeam(ctx, cmd.FromTeamID)
		if err != nil {
			return err
		}
		to, err := tx.GetTeam(ctx, cmd.ToTeamID)
		if err != nil {
			return err
		}
		if from.CurrentIncidentID != dispatch.IncidentID || to.Status != domain.ResourceAvailable {
			return domain.Wrap(domain.ErrConflict, "handoff", "team", cmd.ToTeamID, "team ownership is incompatible", nil)
		}
		from.Status = domain.ResourceAvailable
		from.CurrentIncidentID = ""
		from.UpdatedAt = now
		to.Status = domain.ResourceDeployed
		to.CurrentIncidentID = dispatch.IncidentID
		to.UpdatedAt = now
		dispatch.TeamID = to.ID
		dispatch.UpdatedAt = now
		dispatch.Version++
		if err := tx.UpdateTeam(ctx, from, cmd.FromVersion); err != nil {
			return err
		}
		if err := tx.UpdateTeam(ctx, to, cmd.ToVersion); err != nil {
			return err
		}
		if err := tx.UpdateDispatch(ctx, dispatch, dispatch.Version-1); err != nil {
			return err
		}
		if err := c.outbox(ctx, tx, "team.handed_off", "dispatch", dispatch.ID, dispatch); err != nil {
			return err
		}
		changed = dispatch
		return c.audit(ctx, tx, cmd.Actor, "team.handoff", "dispatch", dispatch.ID, "success", fmt.Sprintf("%s to %s", from.ID, to.ID))
	})
	if err != nil {
		return domain.Dispatch{}, fmt.Errorf("handoff team: %w", err)
	}
	return changed, nil
}
