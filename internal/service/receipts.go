package service

import (
	"context"
	"fmt"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
	"github.com/11DingKing/nanjing-tourism-hub/internal/repository"
)

func (c *Coordinator) SubmitReceipt(ctx context.Context, cmd domain.SubmitReceiptCommand) (domain.Receipt, error) {
	if err := domain.RequireRole(cmd.Actor, domain.RoleFieldLead, domain.RoleDispatcher); err != nil {
		return domain.Receipt{}, err
	}
	if cmd.Receipt.EvidenceCount < 1 {
		return domain.Receipt{}, fmt.Errorf("%w: receipt needs evidence", domain.ErrInvalidState)
	}
	now := c.now().UTC()
	receipt := cmd.Receipt
	if receipt.ID == "" {
		receipt.ID = newID("receipt")
	}
	receipt.ReporterID = cmd.Actor.UserID
	receipt.Status = "submitted"
	receipt.SubmittedAt = now
	receipt.Version = 1
	err := c.repo.WithinTx(ctx, func(tx repository.Tx) error {
		task, err := tx.GetTask(ctx, receipt.TaskID)
		if err != nil {
			return err
		}
		if task.Status != domain.TaskExecuting {
			return domain.Wrap(domain.ErrInvalidState, "submit", "receipt", receipt.ID, "task is not executing", nil)
		}
		if task.IncidentID != receipt.IncidentID || task.RegionID != receipt.RegionID {
			return domain.Wrap(domain.ErrInvalidState, "submit", "receipt", receipt.ID, "receipt scope differs from task", nil)
		}
		if err := tx.InsertReceipt(ctx, receipt); err != nil {
			return err
		}
		task.Status = domain.TaskCompleted
		task.UpdatedAt = now
		if err := tx.UpdateTask(ctx, task, cmd.TaskVersion); err != nil {
			return err
		}
		if err := c.outbox(ctx, tx, "receipt.submitted", "receipt", receipt.ID, receipt); err != nil {
			return err
		}
		return c.audit(ctx, tx, cmd.Actor, "receipt.submit", "receipt", receipt.ID, "success", receipt.Summary)
	})
	if err != nil {
		return domain.Receipt{}, fmt.Errorf("submit receipt: %w", err)
	}
	return receipt, nil
}
func (c *Coordinator) ReviewReceipt(ctx context.Context, cmd domain.ReviewReceiptCommand) (domain.Receipt, error) {
	if err := domain.RequireRole(cmd.Actor, domain.RoleReviewer, domain.RoleDutyOfficer); err != nil {
		return domain.Receipt{}, err
	}
	var changed domain.Receipt
	err := c.repo.WithinTx(ctx, func(tx repository.Tx) error {
		receipt, err := tx.GetReceipt(ctx, cmd.ReceiptID)
		if err != nil {
			return err
		}
		if receipt.Status != "submitted" {
			return domain.Wrap(domain.ErrInvalidState, "review", "receipt", receipt.ID, "receipt is not pending", nil)
		}
		now := c.now().UTC()
		if cmd.Accept {
			receipt.Status = "accepted"
		} else {
			receipt.Status = "rejected"
		}
		receipt.ReviewedAt = &now
		receipt.Version++
		if err := tx.UpdateReceipt(ctx, receipt, cmd.ExpectedVersion); err != nil {
			return err
		}
		if !cmd.Accept {
			task, err := tx.GetTask(ctx, receipt.TaskID)
			if err != nil {
				return err
			}
			task.Status = domain.TaskExecuting
			task.UpdatedAt = now
			if err := tx.UpdateTask(ctx, task, task.Version); err != nil {
				return err
			}
		}
		if err := c.outbox(ctx, tx, "receipt.reviewed", "receipt", receipt.ID, map[string]any{"receipt": receipt, "note": cmd.Note}); err != nil {
			return err
		}
		changed = receipt
		return c.audit(ctx, tx, cmd.Actor, "receipt.review", "receipt", receipt.ID, "success", cmd.Note)
	})
	if err != nil {
		return domain.Receipt{}, fmt.Errorf("review receipt: %w", err)
	}
	return changed, nil
}
