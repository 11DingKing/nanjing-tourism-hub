package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
	"github.com/11DingKing/nanjing-tourism-hub/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type LoginResult struct {
	Token     string
	SessionID string
	ExpiresAt time.Time
	User      domain.User
}

func (c *Coordinator) Login(ctx context.Context, username, password string) (LoginResult, error) {
	user, err := c.repo.FindUserByUsername(ctx, username)
	if err != nil {
		return LoginResult{}, domain.Wrap(domain.ErrUnauthorized, "login", "user", username, "credentials rejected", err)
	}
	if !user.Active || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return LoginResult{}, domain.Wrap(domain.ErrUnauthorized, "login", "user", username, "credentials rejected", nil)
	}
	token := newID("token")
	digest := sha256.Sum256([]byte(token))
	now := c.now().UTC()
	session := domain.Session{ID: newID("session"), UserID: user.ID, TokenHash: hex.EncodeToString(digest[:]), ExpiresAt: now.Add(c.sessionTTL), CreatedAt: now}
	err = c.repo.WithinTx(ctx, func(tx repository.Tx) error {
		if err := tx.CreateSession(ctx, session); err != nil {
			return err
		}
		return c.audit(ctx, tx, domain.Actor{UserID: user.ID, Role: user.Role, RegionID: user.RegionID, RequestID: newID("login")}, "session.login", "session", session.ID, "success", "")
	})
	if err != nil {
		return LoginResult{}, fmt.Errorf("persist login: %w", err)
	}
	user.PasswordHash = ""
	return LoginResult{Token: token, SessionID: session.ID, ExpiresAt: session.ExpiresAt, User: user}, nil
}
func (c *Coordinator) Authenticate(ctx context.Context, token string) (domain.Actor, error) {
	digest := sha256.Sum256([]byte(token))
	session, user, err := c.repo.FindSessionByHash(ctx, hex.EncodeToString(digest[:]))
	if err != nil {
		return domain.Actor{}, domain.Wrap(domain.ErrUnauthorized, "authenticate", "session", "", "token rejected", err)
	}
	now := c.now()
	if session.RevokedAt != nil {
		return domain.Actor{}, domain.Wrap(domain.ErrUnauthorized, "authenticate", "session", session.ID, "session revoked", nil)
	}
	if !session.ExpiresAt.After(now) {
		return domain.Actor{}, domain.Wrap(domain.ErrExpired, "authenticate", "session", session.ID, "session expired", nil)
	}
	if !user.Active {
		return domain.Actor{}, domain.Wrap(domain.ErrUnauthorized, "authenticate", "user", user.ID, "user inactive", nil)
	}
	return domain.Actor{UserID: user.ID, Role: user.Role, RegionID: user.RegionID}, nil
}
func (c *Coordinator) Logout(ctx context.Context, sessionID string, actor domain.Actor) error {
	if err := domain.ValidateActor(actor); err != nil {
		return err
	}
	if err := c.repo.RevokeSession(ctx, sessionID, c.now().UTC()); err != nil {
		return fmt.Errorf("logout: %w", err)
	}
	return nil
}
