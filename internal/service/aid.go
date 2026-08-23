package service

import (
	"context"
	"github.com/11DingKing/autumn-grain-resilience/internal/domain"
	"github.com/11DingKing/autumn-grain-resilience/internal/repository"
)

type AidService struct {
	Repo  repository.AidRepo
	Audit *AuditService
}

func (s *AidService) Create(ctx context.Context, v domain.AidCase, request string) error {
	if v.Status == "" {
		v.Status = "draft"
	}
	if v.Version == 0 {
		v.Version = 1
	}
	if e := s.Repo.Create(ctx, v); e != nil {
		return e
	}
	return s.Audit.Record(ctx, "system", "create_aid_case", "aid_case", v.ID, "success", request)
}
func (s *AidService) Transition(ctx context.Context, id, from, to string, version int, actor, request string) error {
	if !domain.ValidAidTransition(from, to) {
		return domain.ErrInvalidState
	}
	ok, e := s.Repo.Transition(ctx, id, from, to, version)
	if e != nil {
		return e
	}
	if !ok {
		return domain.ErrConflict
	}
	return s.Audit.Record(ctx, actor, "transition_aid_case", "aid_case", id, to, request)
}
