package service

import (
	"context"
	"fmt"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
	"github.com/11DingKing/nanjing-tourism-hub/internal/repository"
)

func (c *Coordinator) PublishWarning(ctx context.Context, cmd domain.PublishWarningCommand) (domain.Warning, error) {
	if err := domain.RequireRole(cmd.Actor, domain.RoleDutyOfficer); err != nil {
		return domain.Warning{}, err
	}
	if err := domain.ValidateWindow(cmd.Warning.EffectiveFrom, cmd.Warning.EffectiveUntil); err != nil {
		return domain.Warning{}, err
	}
	if cmd.Warning.RainfallMM < 0 {
		return domain.Warning{}, fmt.Errorf("%w: rainfall cannot be negative", domain.ErrInvalidState)
	}
	v := cmd.Warning
	if v.ID == "" {
		v.ID = newID("warning")
	}
	v.Version = 1
	v.Status = "published"
	err := c.repo.WithinTx(ctx, func(tx repository.Tx) error {
		if err := tx.InsertWarning(ctx, v); err != nil {
			return err
		}
		if err := c.outbox(ctx, tx, "warning.published", "warning", v.ID, v); err != nil {
			return err
		}
		return c.audit(ctx, tx, cmd.Actor, "warning.publish", "warning", v.ID, "success", v.LevelString())
	})
	if err != nil {
		return domain.Warning{}, fmt.Errorf("publish warning: %w", err)
	}
	return v, nil
}
func (c *Coordinator) ReviseWarning(ctx context.Context, id string, expected int64, next domain.Warning, actor domain.Actor) (domain.Warning, error) {
	if err := domain.RequireRole(actor, domain.RoleDutyOfficer); err != nil {
		return domain.Warning{}, err
	}
	if err := domain.ValidateWindow(next.EffectiveFrom, next.EffectiveUntil); err != nil {
		return domain.Warning{}, err
	}
	err := c.repo.WithinTx(ctx, func(tx repository.Tx) error {
		current, err := tx.GetWarning(ctx, id)
		if err != nil {
			return err
		}
		next.ID = current.ID
		next.Number = current.Number
		next.TyphoonName = current.TyphoonName
		next.Version = current.Version + 1
		if err := tx.UpdateWarning(ctx, next, expected); err != nil {
			return err
		}
		if err := c.outbox(ctx, tx, "warning.revised", "warning", id, next); err != nil {
			return err
		}
		return c.audit(ctx, tx, actor, "warning.revise", "warning", id, "success", string(next.Level))
	})
	if err != nil {
		return domain.Warning{}, fmt.Errorf("revise warning: %w", err)
	}
	return next, nil
}
