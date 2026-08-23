package service

import (
	"context"
	"fmt"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
	"github.com/11DingKing/nanjing-tourism-hub/internal/repository"
)

func (c *Coordinator) ReserveShelter(ctx context.Context, cmd domain.ReserveShelterCommand) (domain.Reservation, error) {
	if err := domain.RequireRole(cmd.Actor, domain.RoleDispatcher, domain.RoleDutyOfficer); err != nil {
		return domain.Reservation{}, err
	}
	if err := requireRegion(cmd.Actor, cmd.RegionID); err != nil {
		return domain.Reservation{}, err
	}
	if cmd.People <= 0 {
		return domain.Reservation{}, fmt.Errorf("%w: people must be positive", domain.ErrInvalidState)
	}
	if !cmd.ExpiresAt.After(c.now()) {
		return domain.Reservation{}, fmt.Errorf("%w: reservation must expire in future", domain.ErrExpired)
	}
	now := c.now().UTC()
	reservation := domain.Reservation{ID: cmd.ReservationID, ShelterID: cmd.ShelterID, IncidentID: cmd.IncidentID, RegionID: cmd.RegionID, People: cmd.People, Status: "reserved", IdempotencyKey: cmd.IdempotencyKey, ExpiresAt: cmd.ExpiresAt.UTC(), Version: 1, CreatedAt: now, UpdatedAt: now}
	if reservation.ID == "" {
		reservation.ID = newID("reservation")
	}
	err := c.repo.WithinTx(ctx, func(tx repository.Tx) error {
		shelter, err := tx.GetShelter(ctx, cmd.ShelterID)
		if err != nil {
			return err
		}
		if shelter.RegionID != cmd.RegionID || shelter.Status == domain.ResourceMaintenance {
			return domain.Wrap(domain.ErrInvalidState, "reserve", "shelter", shelter.ID, "shelter cannot accept reservation", nil)
		}
		if shelter.Capacity-shelter.Reserved-shelter.Occupied < cmd.People {
			return domain.Wrap(domain.ErrCapacity, "reserve", "shelter", shelter.ID, "insufficient remaining capacity", nil)
		}
		shelter.Reserved += cmd.People
		shelter.UpdatedAt = now
		if err := tx.UpdateShelter(ctx, shelter, shelter.Version); err != nil {
			return err
		}
		if err := tx.InsertReservation(ctx, reservation); err != nil {
			return err
		}
		if err := c.job(ctx, tx, "reservation.expire", reservation.ID, map[string]string{"reservation_id": reservation.ID}, reservation.ExpiresAt); err != nil {
			return err
		}
		return c.audit(ctx, tx, cmd.Actor, "shelter.reserve", "reservation", reservation.ID, "success", fmt.Sprintf("people=%d", cmd.People))
	})
	if err != nil {
		return domain.Reservation{}, fmt.Errorf("reserve shelter: %w", err)
	}
	return reservation, nil
}
func (c *Coordinator) ReleaseShelter(ctx context.Context, cmd domain.ReleaseShelterCommand) error {
	if err := domain.RequireRole(cmd.Actor, domain.RoleDispatcher, domain.RoleDutyOfficer); err != nil {
		return err
	}
	now := c.now().UTC()
	return c.repo.WithinTx(ctx, func(tx repository.Tx) error {
		reservation, err := tx.GetReservation(ctx, cmd.ReservationID)
		if err != nil {
			return err
		}
		if reservation.Status != "reserved" {
			return domain.Wrap(domain.ErrInvalidState, "release", "reservation", reservation.ID, "reservation is not active", nil)
		}
		shelter, err := tx.GetShelter(ctx, reservation.ShelterID)
		if err != nil {
			return err
		}
		if shelter.Reserved < reservation.People {
			return domain.Wrap(domain.ErrInvalidState, "release", "shelter", shelter.ID, "reserved capacity is inconsistent", nil)
		}
		shelter.Reserved -= reservation.People
		shelter.UpdatedAt = now
		reservation.Status = "released"
		reservation.UpdatedAt = now
		if err := tx.UpdateShelter(ctx, shelter, shelter.Version); err != nil {
			return err
		}
		if err := tx.UpdateReservation(ctx, reservation, cmd.ExpectedVersion); err != nil {
			return err
		}
		return c.audit(ctx, tx, cmd.Actor, "shelter.release", "reservation", reservation.ID, "success", "")
	})
}
