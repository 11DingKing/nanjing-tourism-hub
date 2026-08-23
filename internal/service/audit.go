package service

import (
	"context"
	"github.com/11DingKing/autumn-grain-resilience/internal/domain"
	"github.com/11DingKing/autumn-grain-resilience/internal/repository"
	"log/slog"
	"time"
)

type AuditService struct {
	Repo   repository.AuditRepo
	Logger *slog.Logger
}

func (s *AuditService) Record(ctx context.Context, actor, action, typ, obj, result, request string) error {
	e := domain.AuditEvent{ID: id("aud_"), ActorID: actor, Action: action, ObjectType: typ, ObjectID: obj, Result: result, RequestID: request, CreatedAt: time.Now().UTC()}
	if err := s.Repo.Append(ctx, e); err != nil {
		s.Logger.Error("audit append", "error", err)
		return err
	}
	return nil
}
