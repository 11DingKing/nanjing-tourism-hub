package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/11DingKing/autumn-grain-resilience/internal/domain"
	"time"
)

type UserRepo struct{ DB DB }

func (r UserRepo) Create(ctx context.Context, u domain.User) error {
	_, err := r.DB.ExecContext(ctx, "INSERT INTO users(id,name,role,password_hash,disabled,created_at) VALUES(?,?,?,?,?,?)", u.ID, u.Name, u.Role, u.PasswordHash, u.Disabled, fmtTime(u.CreatedAt))
	return err
}
func (r UserRepo) ByID(ctx context.Context, id string) (domain.User, error) {
	var u domain.User
	var created string
	var disabled int
	err := r.DB.QueryRowContext(ctx, "SELECT id,name,role,password_hash,disabled,created_at FROM users WHERE id=?", id).Scan(&u.ID, &u.Name, &u.Role, &u.PasswordHash, &disabled, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return u, domain.ErrNotFound
	}
	u.Disabled = disabled != 0
	u.CreatedAt = parseTime(created)
	return u, err
}
func (r UserRepo) CreateSession(ctx context.Context, s domain.Session) error {
	_, err := r.DB.ExecContext(ctx, "INSERT INTO sessions(id,user_id,expires_at) VALUES(?,?,?)", s.ID, s.UserID, fmtTime(s.ExpiresAt))
	return err
}
func (r UserRepo) Session(ctx context.Context, id string) (domain.Session, error) {
	var s domain.Session
	var exp string
	var revoked sql.NullString
	err := r.DB.QueryRowContext(ctx, "SELECT id,user_id,expires_at,revoked_at FROM sessions WHERE id=?", id).Scan(&s.ID, &s.UserID, &exp, &revoked)
	if errors.Is(err, sql.ErrNoRows) {
		return s, domain.ErrNotFound
	}
	s.ExpiresAt = parseTime(exp)
	if revoked.Valid {
		t := parseTime(revoked.String)
		s.RevokedAt = &t
	}
	return s, err
}
func (r UserRepo) RevokeSession(ctx context.Context, id string, t time.Time) error {
	_, err := r.DB.ExecContext(ctx, "UPDATE sessions SET revoked_at=? WHERE id=?", fmtTime(t), id)
	return err
}
