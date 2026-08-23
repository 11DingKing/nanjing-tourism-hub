package service

import (
	"context"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
	"github.com/11DingKing/nanjing-tourism-hub/internal/repository"
)

func (c *Coordinator) BatchAssign(ctx context.Context, cmd domain.BatchAssignCommand) (domain.BatchResult, error) {
	if err := domain.RequireRole(cmd.Actor, domain.RoleDispatcher, domain.RoleDutyOfficer); err != nil {
		return domain.BatchResult{}, err
	}
	result := domain.BatchResult{Items: make([]domain.BatchItemResult, 0, len(cmd.Items))}
	for _, item := range cmd.Items {
		if err := ctx.Err(); err != nil {
			return result, domain.Wrap(domain.ErrCanceled, "batch_assign", "incident", cmd.IncidentID, "request canceled", err)
		}
		entry := domain.BatchItemResult{InputID: item.TaskID, ResourceID: item.TeamID}
		err := c.repo.WithinTx(ctx, func(tx repository.Tx) error {
			task, err := tx.GetTask(ctx, item.TaskID)
			if err != nil {
				return err
			}
			team, err := tx.GetTeam(ctx, item.TeamID)
			if err != nil {
				return err
			}
			if task.IncidentID != cmd.IncidentID || task.Status != domain.TaskPending {
				return domain.Wrap(domain.ErrInvalidState, "batch_assign", "task", task.ID, "task is not assignable", nil)
			}
			if team.Status != domain.ResourceAvailable {
				return domain.Wrap(domain.ErrConflict, "batch_assign", "team", team.ID, "team is unavailable", nil)
			}
			now := c.now().UTC()
			task.Status = domain.TaskAccepted
			task.AssigneeID = team.ID
			task.UpdatedAt = now
			team.Status = domain.ResourceDeployed
			team.CurrentIncidentID = cmd.IncidentID
			team.UpdatedAt = now
			if err := tx.UpdateTask(ctx, task, item.TaskVersion); err != nil {
				return err
			}
			if err := tx.UpdateTeam(ctx, team, item.TeamVersion); err != nil {
				return err
			}
			dispatch := domain.Dispatch{ID: newID("dispatch"), TeamID: team.ID, IncidentID: cmd.IncidentID, RegionID: task.RegionID, TaskID: task.ID, Status: domain.TaskAccepted, RequestedBy: cmd.Actor.UserID, Version: 1, CreatedAt: now, UpdatedAt: now}
			if err := tx.InsertDispatch(ctx, dispatch); err != nil {
				return err
			}
			entry.Status = "assigned"
			return c.audit(ctx, tx, cmd.Actor, "batch.assign", "dispatch", dispatch.ID, "success", "")
		})
		if err != nil {
			entry.Status = "failed"
			entry.ErrorCode = "assignment_failed"
			result.Failed++
		} else {
			result.Succeeded++
		}
		result.Items = append(result.Items, entry)
	}
	return result, nil
}
