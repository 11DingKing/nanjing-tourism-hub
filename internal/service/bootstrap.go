package service

import (
	"context"
	"errors"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
	"github.com/11DingKing/nanjing-tourism-hub/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type BootstrapInput struct {
	Region     domain.Region
	Users      []BootstrapUser
	Shelters   []domain.Shelter
	Teams      []domain.Team
	SupplyLots []domain.SupplyLot
}
type BootstrapUser struct {
	ID, Username, Password string
	Role                   domain.Role
	RegionID               string
}

func (c *Coordinator) Bootstrap(ctx context.Context, input BootstrapInput) error {
	now := c.now().UTC()
	return c.repo.WithinTx(ctx, func(tx repository.Tx) error {
		if err := tx.InsertRegion(ctx, input.Region); err != nil && !errors.Is(err, domain.ErrConflict) {
			return err
		}
		for _, seed := range input.Users {
			hash, err := bcrypt.GenerateFromPassword([]byte(seed.Password), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			user := domain.User{ID: seed.ID, Username: seed.Username, PasswordHash: string(hash), Role: seed.Role, RegionID: seed.RegionID, Active: true, Version: 1, CreatedAt: now, UpdatedAt: now}
			if err := tx.CreateUser(ctx, user); err != nil && !errors.Is(err, domain.ErrConflict) {
				return err
			}
		}
		for _, v := range input.Shelters {
			if v.Version == 0 {
				v.Version = 1
			}
			v.UpdatedAt = now
			if err := tx.InsertShelter(ctx, v); err != nil && !errors.Is(err, domain.ErrConflict) {
				return err
			}
		}
		for _, v := range input.Teams {
			if v.Version == 0 {
				v.Version = 1
			}
			v.UpdatedAt = now
			if err := tx.InsertTeam(ctx, v); err != nil && !errors.Is(err, domain.ErrConflict) {
				return err
			}
		}
		for _, v := range input.SupplyLots {
			if v.Version == 0 {
				v.Version = 1
			}
			v.UpdatedAt = now
			if v.ExpiresAt.IsZero() {
				v.ExpiresAt = now.Add(30 * 24 * time.Hour)
			}
			if err := tx.InsertSupplyLot(ctx, v); err != nil && !errors.Is(err, domain.ErrConflict) {
				return err
			}
		}
		return nil
	})
}
