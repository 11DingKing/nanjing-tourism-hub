package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
	"github.com/11DingKing/nanjing-tourism-hub/internal/repository"
)

type Clock func() time.Time
type Coordinator struct {
	repo       repository.Repository
	now        Clock
	sessionTTL time.Duration
}

func New(repo repository.Repository, now Clock, sessionTTL time.Duration) *Coordinator {
	if now == nil {
		now = time.Now
	}
	return &Coordinator{repo: repo, now: now, sessionTTL: sessionTTL}
}
func newID(prefix string) string {
	var raw [12]byte
	if _, err := rand.Read(raw[:]); err != nil {
		panic(err)
	}
	return prefix + "_" + hex.EncodeToString(raw[:])
}
func (c *Coordinator) audit(ctx context.Context, tx repository.Tx, actor domain.Actor, action, entity, id, result, detail string) error {
	return tx.InsertAudit(ctx, domain.AuditEvent{ID: newID("audit"), ActorID: actor.UserID, RequestID: actor.RequestID, Action: action, Entity: entity, EntityID: id, Result: result, Detail: detail, CreatedAt: c.now().UTC()})
}
func (c *Coordinator) outbox(ctx context.Context, tx repository.Tx, topic, kind, id string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode outbox %s: %w", topic, err)
	}
	now := c.now().UTC()
	return tx.InsertOutbox(ctx, domain.OutboxEvent{ID: newID("evt"), Topic: topic, AggregateType: kind, AggregateID: id, Payload: body, Status: "pending", NextAttemptAt: now, CreatedAt: now, UpdatedAt: now})
}
func (c *Coordinator) job(ctx context.Context, tx repository.Tx, kind, id string, payload any, runAt time.Time) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode job %s: %w", kind, err)
	}
	now := c.now().UTC()
	return tx.InsertJob(ctx, domain.Job{ID: newID("job"), Kind: kind, AggregateID: id, Payload: body, Status: "pending", MaxAttempts: 5, NextRunAt: runAt.UTC(), CreatedAt: now, UpdatedAt: now})
}
func requireRegion(actor domain.Actor, regionID string) error {
	if actor.Role == domain.RoleDutyOfficer || actor.Role == domain.RoleReviewer {
		return nil
	}
	if actor.RegionID != regionID {
		return domain.Wrap(domain.ErrForbidden, "authorize", "region", regionID, "actor belongs to another region", nil)
	}
	return nil
}
