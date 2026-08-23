package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
)

func (t *txStore) CreateUser(ctx context.Context, value domain.User) error {
	_, err := t.tx.ExecContext(ctx, `INSERT INTO users(id,username,password_hash,role,region_id,active,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?)`,
		value.ID, value.Username, value.PasswordHash, value.Role, nullText(value.RegionID), boolInt(value.Active), value.Version, stamp(value.CreatedAt), stamp(value.UpdatedAt))
	return translate("create", "user", value.ID, err)
}

func (t *txStore) CreateSession(ctx context.Context, value domain.Session) error {
	_, err := t.tx.ExecContext(ctx, `INSERT INTO sessions(id,user_id,token_hash,expires_at,revoked_at,created_at) VALUES(?,?,?,?,?,?)`,
		value.ID, value.UserID, value.TokenHash, stamp(value.ExpiresAt), nullableStamp(value.RevokedAt), stamp(value.CreatedAt))
	return translate("create", "session", value.ID, err)
}

func (t *txStore) RevokeSessionsForUser(ctx context.Context, userID string, at time.Time) error {
	_, err := t.tx.ExecContext(ctx, `UPDATE sessions SET revoked_at=? WHERE user_id=? AND revoked_at IS NULL`, stamp(at), userID)
	return translate("revoke", "session", userID, err)
}

func (s *Store) FindSessionByHash(ctx context.Context, tokenHash string) (domain.Session, domain.User, error) {
	var session domain.Session
	var user domain.User
	var revoked sql.NullString
	var region sql.NullString
	var expires, sessionCreated, userCreated, userUpdated string
	var active int
	err := s.db.QueryRowContext(ctx, `SELECT s.id,s.user_id,s.token_hash,s.expires_at,s.revoked_at,s.created_at,u.id,u.username,u.password_hash,u.role,u.region_id,u.active,u.version,u.created_at,u.updated_at FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.token_hash=?`, tokenHash).Scan(
		&session.ID, &session.UserID, &session.TokenHash, &expires, &revoked, &sessionCreated, &user.ID, &user.Username, &user.PasswordHash, &user.Role, &region, &active, &user.Version, &userCreated, &userUpdated)
	if err != nil {
		return domain.Session{}, domain.User{}, translate("find", "session", tokenHash, err)
	}
	session.ExpiresAt = parseStamp(expires)
	session.RevokedAt = scanOptional(revoked)
	session.CreatedAt = parseStamp(sessionCreated)
	user.RegionID = region.String
	user.Active = active == 1
	user.CreatedAt = parseStamp(userCreated)
	user.UpdatedAt = parseStamp(userUpdated)
	return session, user, nil
}

func (s *Store) RevokeSession(ctx context.Context, id string, at time.Time) error {
	result, err := s.db.ExecContext(ctx, `UPDATE sessions SET revoked_at=? WHERE id=? AND revoked_at IS NULL`, stamp(at), id)
	if err != nil {
		return translate("revoke", "session", id, err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return domain.Wrap(domain.ErrNotFound, "revoke", "session", id, "active session not found", nil)
	}
	return nil
}

func (s *Store) FindUserByUsername(ctx context.Context, username string) (domain.User, error) {
	var user domain.User
	var region sql.NullString
	var active int
	var created, updated string
	err := s.db.QueryRowContext(ctx, `SELECT id,username,password_hash,role,region_id,active,version,created_at,updated_at FROM users WHERE username=?`, username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &region, &active, &user.Version, &created, &updated)
	if err != nil {
		return domain.User{}, translate("find", "user", username, err)
	}
	user.RegionID = region.String
	user.Active = active == 1
	user.CreatedAt = parseStamp(created)
	user.UpdatedAt = parseStamp(updated)
	return user, nil
}

func nullText(value string) any {
	if value == "" {
		return nil
	}
	return value
}
