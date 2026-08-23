package service

import (
	"context"
	"github.com/11DingKing/autumn-grain-resilience/internal/domain"
	"github.com/11DingKing/autumn-grain-resilience/internal/repository"
	"time"
)

type FieldTaskService struct {
	Repo  repository.TaskRepo
	Plots repository.PlotRepo
	Audit *AuditService
}

func (s *FieldTaskService) Create(ctx context.Context, v domain.FieldTask, request string) error {
	if v.Status == "" {
		v.Status = "queued"
	}
	if v.Version == 0 {
		v.Version = 1
	}
	if _, e := s.Plots.ByID(ctx, v.PlotID); e != nil {
		return e
	}
	if e := s.Repo.Create(ctx, v); e != nil {
		return e
	}
	return s.Audit.Record(ctx, v.AssigneeID, "create_field_task", "field_task", v.ID, "success", request)
}
func (s *FieldTaskService) Claim(ctx context.Context, taskID, worker, request string) (bool, error) {
	ok, e := s.Repo.Claim(ctx, taskID, worker, id("lease_"), time.Now().Add(15*time.Minute))
	if e != nil {
		return false, e
	}
	if ok {
		e = s.Audit.Record(ctx, worker, "claim_field_task", "field_task", taskID, "success", request)
	}
	return ok, e
}
func (s *FieldTaskService) Complete(ctx context.Context, id, request string) error {
	v, e := s.Repo.ByID(ctx, id)
	if e != nil {
		return e
	}
	if !domain.ValidTaskTransition(v.Status, "in_progress") && v.Status != "in_progress" {
		return domain.ErrInvalidState
	}
	ok, e := s.Repo.Transition(ctx, id, v.Status, "completed", v.Version)
	if e != nil {
		return e
	}
	if !ok {
		return domain.ErrConflict
	}
	return s.Audit.Record(ctx, v.AssigneeID, "complete_field_task", "field_task", id, "success", request)
}
