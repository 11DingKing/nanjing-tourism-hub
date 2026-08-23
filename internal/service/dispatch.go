package service

import (
	"context"
	"fmt"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
	"github.com/11DingKing/nanjing-tourism-hub/internal/repository"
)

func (c *Coordinator) DispatchTeam(ctx context.Context, cmd domain.DispatchTeamCommand) (domain.Dispatch, error) {
	if err := domain.RequireRole(cmd.Actor, domain.RoleDispatcher, domain.RoleDutyOfficer); err != nil {
		return domain.Dispatch{}, err
	}
	if err := requireRegion(cmd.Actor, cmd.RegionID); err != nil {
		return domain.Dispatch{}, err
	}
	now := c.now().UTC()
	dispatch := domain.Dispatch{ID: cmd.DispatchID, TeamID: cmd.TeamID, IncidentID: cmd.IncidentID, RegionID: cmd.RegionID, TaskID: cmd.TaskID, Status: domain.TaskPending, RequestedBy: cmd.Actor.UserID, Version: 1, CreatedAt: now, UpdatedAt: now}
	if dispatch.ID == "" {
		dispatch.ID = newID("dispatch")
	}
	err := c.repo.WithinTx(ctx, func(tx repository.Tx) error {
		team, err := tx.GetTeam(ctx, cmd.TeamID)
		if err != nil {
			return err
		}
		task, err := tx.GetTask(ctx, cmd.TaskID)
		if err != nil {
			return err
		}
		if team.Status != domain.ResourceAvailable || team.CurrentIncidentID != "" {
			return domain.Wrap(domain.ErrConflict, "dispatch", "team", team.ID, "team is not available", nil)
		}
		if task.IncidentID != cmd.IncidentID || task.RegionID != cmd.RegionID || task.Status != domain.TaskPending {
			return domain.Wrap(domain.ErrInvalidState, "dispatch", "task", task.ID, "task cannot be assigned", nil)
		}
		team.Status = domain.ResourceDeployed
		team.CurrentIncidentID = cmd.IncidentID
		team.UpdatedAt = now
		task.Status = domain.TaskAccepted
		task.AssigneeID = team.ID
		task.UpdatedAt = now
		if err := tx.UpdateTeam(ctx, team, cmd.TeamVersion); err != nil {
			return err
		}
		if err := tx.UpdateTask(ctx, task, task.Version); err != nil {
			return err
		}
		if err := tx.InsertDispatch(ctx, dispatch); err != nil {
			return err
		}
		if err := c.outbox(ctx, tx, "team.dispatched", "dispatch", dispatch.ID, dispatch); err != nil {
			return err
		}
		return c.audit(ctx, tx, cmd.Actor, "team.dispatch", "dispatch", dispatch.ID, "success", team.Name)
	})
	if err != nil {
		return domain.Dispatch{}, fmt.Errorf("dispatch team: %w", err)
	}
	return dispatch, nil
}
