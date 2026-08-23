package service

import (
	"context"
	"fmt"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
	"github.com/11DingKing/nanjing-tourism-hub/internal/repository"
)

func (c *Coordinator) CreateTask(ctx context.Context, cmd domain.CreateTaskCommand) (domain.DistrictTask, error) {
	if err := domain.RequireRole(cmd.Actor, domain.RoleDutyOfficer, domain.RoleDispatcher); err != nil {
		return domain.DistrictTask{}, err
	}
	if err := requireRegion(cmd.Actor, cmd.Task.RegionID); err != nil {
		return domain.DistrictTask{}, err
	}
	if !cmd.Task.Deadline.After(c.now()) {
		return domain.DistrictTask{}, fmt.Errorf("%w: task deadline must be in future", domain.ErrExpired)
	}
	now := c.now().UTC()
	task := cmd.Task
	if task.ID == "" {
		task.ID = newID("task")
	}
	task.Status = domain.TaskPending
	task.CreatedBy = cmd.Actor.UserID
	task.Version = 1
	task.CreatedAt = now
	task.UpdatedAt = now
	err := c.repo.WithinTx(ctx, func(tx repository.Tx) error {
		incident, err := tx.GetIncident(ctx, task.IncidentID)
		if err != nil {
			return err
		}
		if incident.Status != domain.IncidentActive {
			return domain.Wrap(domain.ErrInvalidState, "create", "task", task.ID, "incident is not active", nil)
		}
		if _, err := tx.GetResponse(ctx, task.IncidentID, task.RegionID); err != nil {
			return err
		}
		if err := tx.InsertTask(ctx, task); err != nil {
			return err
		}
		if err := c.outbox(ctx, tx, "task.created", "task", task.ID, task); err != nil {
			return err
		}
		return c.audit(ctx, tx, cmd.Actor, "task.create", "task", task.ID, "success", task.Kind)
	})
	if err != nil {
		return domain.DistrictTask{}, fmt.Errorf("create task: %w", err)
	}
	return task, nil
}
func (c *Coordinator) AdvanceTask(ctx context.Context, cmd domain.AdvanceTaskCommand) (domain.DistrictTask, error) {
	if err := domain.RequireRole(cmd.Actor, domain.RoleFieldLead, domain.RoleDispatcher, domain.RoleDutyOfficer); err != nil {
		return domain.DistrictTask{}, err
	}
	var changed domain.DistrictTask
	err := c.repo.WithinTx(ctx, func(tx repository.Tx) error {
		task, err := tx.GetTask(ctx, cmd.TaskID)
		if err != nil {
			return err
		}
		if err := requireRegion(cmd.Actor, task.RegionID); err != nil {
			return err
		}
		if err := domain.MoveTask(task.Status, cmd.Target); err != nil {
			return err
		}
		if cmd.Actor.Role == domain.RoleFieldLead && task.AssigneeID != cmd.Actor.UserID {
			return domain.Wrap(domain.ErrForbidden, "advance", "task", task.ID, "field lead does not own task", nil)
		}
		task.Status = cmd.Target
		task.Version++
		task.UpdatedAt = c.now().UTC()
		if err := tx.UpdateTask(ctx, task, cmd.ExpectedVersion); err != nil {
			return err
		}
		if err := c.outbox(ctx, tx, "task.advanced", "task", task.ID, task); err != nil {
			return err
		}
		changed = task
		return c.audit(ctx, tx, cmd.Actor, "task.advance", "task", task.ID, "success", string(cmd.Target))
	})
	if err != nil {
		return domain.DistrictTask{}, fmt.Errorf("advance task: %w", err)
	}
	return changed, nil
}
