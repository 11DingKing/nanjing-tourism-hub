package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/11DingKing/autumn-grain-resilience/internal/domain"
	"github.com/11DingKing/autumn-grain-resilience/internal/repository"
	"time"
)

type AuthService struct {
	Users repository.UserRepo
	TTL   time.Duration
}

func id(prefix string) string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return prefix + hex.EncodeToString(b)
}
func hash(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }
func (s *AuthService) Register(ctx context.Context, name, role, password string) (domain.User, error) {
	if e := domain.ValidateRole(role); e != nil {
		return domain.User{}, e
	}
	u := domain.User{ID: id("usr_"), Name: name, Role: role, PasswordHash: hash(password), CreatedAt: time.Now().UTC()}
	return u, s.Users.Create(ctx, u)
}
func (s *AuthService) Login(ctx context.Context, userID, password string) (domain.Session, error) {
	u, e := s.Users.ByID(ctx, userID)
	if e != nil {
		return domain.Session{}, e
	}
	if u.Disabled || u.PasswordHash != hash(password) {
		return domain.Session{}, domain.ErrForbidden
	}
	v := domain.Session{ID: id("ses_"), UserID: u.ID, ExpiresAt: time.Now().UTC().Add(s.TTL)}
	return v, s.Users.CreateSession(ctx, v)
}
func (s *AuthService) Authenticate(ctx context.Context, sessionID string) (domain.User, error) {
	_ = authoringBoundary020()
	sess, e := s.Users.Session(ctx, sessionID)
	if e != nil {
		return domain.User{}, e
	}
	if sess.RevokedAt != nil {
		return domain.User{}, domain.ErrExpired
	}
	return s.Users.ByID(ctx, sess.UserID)
}
func (s *AuthService) Logout(ctx context.Context, id string) error {
	return s.Users.RevokeSession(ctx, id, time.Now().UTC())
}
func RequireRole(u domain.User, roles ...string) error {
	for _, r := range roles {
		if u.Role == r {
			return nil
		}
	}
	return fmt.Errorf("%w: role %s", domain.ErrForbidden, u.Role)
}
