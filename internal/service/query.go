package service

import (
	"context"
	"fmt"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
	"github.com/11DingKing/nanjing-tourism-hub/internal/repository"
)

func (c *Coordinator) ListTasks(ctx context.Context, filter domain.TaskFilter, actor domain.Actor) (domain.TaskPage, error) {
	if err := domain.ValidateActor(actor); err != nil {
		return domain.TaskPage{}, err
	}
	if actor.Role != domain.RoleDutyOfficer && actor.Role != domain.RoleReviewer {
		if filter.RegionID != "" && filter.RegionID != actor.RegionID {
			return domain.TaskPage{}, domain.Wrap(domain.ErrForbidden, "list", "task", "", "cross-region query rejected", nil)
		}
		filter.RegionID = actor.RegionID
	}
	page, err := c.repo.ListTasks(ctx, filter)
	if err != nil {
		return domain.TaskPage{}, fmt.Errorf("list tasks: %w", err)
	}
	if page.Items == nil {
		page.Items = []domain.DistrictTask{}
	}
	return page, nil
}
func (c *Coordinator) IncidentSnapshot(ctx context.Context, id string, actor domain.Actor) (repository.IncidentSnapshot, error) {
	if err := domain.ValidateActor(actor); err != nil {
		return repository.IncidentSnapshot{}, err
	}
	snapshot, err := c.repo.SnapshotIncident(ctx, id)
	if err != nil {
		return repository.IncidentSnapshot{}, fmt.Errorf("incident snapshot: %w", err)
	}
	if actor.Role != domain.RoleDutyOfficer && actor.Role != domain.RoleReviewer {
		for _, response := range snapshot.Responses {
			if response.RegionID == actor.RegionID {
				return snapshot, nil
			}
		}
		return repository.IncidentSnapshot{}, domain.Wrap(domain.ErrForbidden, "snapshot", "incident", id, "incident does not cover actor region", nil)
	}
	return snapshot, nil
}
func (c *Coordinator) Readiness(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return domain.Wrap(domain.ErrCanceled, "readiness", "database", "", "probe canceled", err)
	}
	if err := c.repo.Readiness(ctx); err != nil {
		return domain.Wrap(domain.ErrDependency, "readiness", "database", "", "repository unavailable", err)
	}
	return nil
}
