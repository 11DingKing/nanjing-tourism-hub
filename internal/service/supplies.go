package service

import (
	"context"
	"fmt"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
	"github.com/11DingKing/nanjing-tourism-hub/internal/repository"
)

func (c *Coordinator) AllocateSupply(ctx context.Context, cmd domain.AllocateSupplyCommand) (domain.Allocation, error) {
	if err := domain.RequireRole(cmd.Actor, domain.RoleDispatcher, domain.RoleDutyOfficer); err != nil {
		return domain.Allocation{}, err
	}
	if err := requireRegion(cmd.Actor, cmd.RegionID); err != nil {
		return domain.Allocation{}, err
	}
	if cmd.Quantity <= 0 {
		return domain.Allocation{}, fmt.Errorf("%w: quantity must be positive", domain.ErrInvalidState)
	}
	now := c.now().UTC()
	allocation := domain.Allocation{ID: cmd.AllocationID, LotID: cmd.LotID, IncidentID: cmd.IncidentID, RegionID: cmd.RegionID, TaskID: cmd.TaskID, Quantity: cmd.Quantity, Status: "reserved", IdempotencyKey: cmd.IdempotencyKey, Version: 1, CreatedAt: now, UpdatedAt: now}
	if allocation.ID == "" {
		allocation.ID = newID("allocation")
	}
	err := c.repo.WithinTx(ctx, func(tx repository.Tx) error {
		lot, err := tx.GetSupplyLot(ctx, cmd.LotID)
		if err != nil {
			return err
		}
		task, err := tx.GetTask(ctx, cmd.TaskID)
		if err != nil {
			return err
		}
		if lot.RegionID != cmd.RegionID || task.IncidentID != cmd.IncidentID {
			return domain.Wrap(domain.ErrInvalidState, "allocate", "supply_lot", lot.ID, "lot and task scope differ", nil)
		}
		if !lot.ExpiresAt.After(now) || lot.Status == domain.ResourceMaintenance {
			return domain.Wrap(domain.ErrExpired, "allocate", "supply_lot", lot.ID, "lot is unavailable", nil)
		}
		if lot.Quantity-lot.Reserved-lot.Issued < cmd.Quantity {
			return domain.Wrap(domain.ErrCapacity, "allocate", "supply_lot", lot.ID, "insufficient available quantity", nil)
		}
		lot.Reserved += cmd.Quantity
		lot.UpdatedAt = now
		if err := tx.UpdateSupplyLot(ctx, lot, cmd.LotVersion); err != nil {
			return err
		}
		if err := tx.InsertAllocation(ctx, allocation); err != nil {
			return err
		}
		if err := c.outbox(ctx, tx, "supply.allocated", "allocation", allocation.ID, allocation); err != nil {
			return err
		}
		return c.audit(ctx, tx, cmd.Actor, "supply.allocate", "allocation", allocation.ID, "success", fmt.Sprintf("quantity=%d", cmd.Quantity))
	})
	if err != nil {
		return domain.Allocation{}, fmt.Errorf("allocate supply: %w", err)
	}
	return allocation, nil
}
func (c *Coordinator) CancelAllocation(ctx context.Context, cmd domain.CancelAllocationCommand) error {
	if err := domain.RequireRole(cmd.Actor, domain.RoleDispatcher, domain.RoleDutyOfficer); err != nil {
		return err
	}
	now := c.now().UTC()
	return c.repo.WithinTx(ctx, func(tx repository.Tx) error {
		allocation, err := tx.GetAllocation(ctx, cmd.AllocationID)
		if err != nil {
			return err
		}
		if allocation.Status != "reserved" {
			return domain.Wrap(domain.ErrInvalidState, "cancel", "allocation", allocation.ID, "only reserved allocation can be canceled", nil)
		}
		lot, err := tx.GetSupplyLot(ctx, allocation.LotID)
		if err != nil {
			return err
		}
		if lot.Reserved < allocation.Quantity {
			return domain.Wrap(domain.ErrInvalidState, "cancel", "supply_lot", lot.ID, "reserved quantity is inconsistent", nil)
		}
		lot.Reserved -= allocation.Quantity
		lot.UpdatedAt = now
		allocation.Status = "canceled"
		allocation.UpdatedAt = now
		if err := tx.UpdateSupplyLot(ctx, lot, lot.Version); err != nil {
			return err
		}
		if err := tx.UpdateAllocation(ctx, allocation, cmd.AllocationVersion); err != nil {
			return err
		}
		return c.audit(ctx, tx, cmd.Actor, "supply.cancel", "allocation", allocation.ID, "success", "")
	})
}
