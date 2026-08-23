package worker

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/11DingKing/nanjing-tourism-hub/internal/repository"
)

type Runner struct {
	repo     repository.Repository
	jobs     JobHandler
	events   EventPublisher
	interval time.Duration
	batch    int
	logger   *slog.Logger
	wg       sync.WaitGroup
}

type JobHandler interface {
	HandleJob(context.Context, string, string, []byte) error
}
type EventPublisher interface {
	Publish(context.Context, string, string, []byte) error
}

func New(repo repository.Repository, jobs JobHandler, events EventPublisher, interval time.Duration, batch int, logger *slog.Logger) *Runner {
	if logger == nil {
		logger = slog.Default()
	}
	if interval <= 0 {
		interval = time.Second
	}
	if batch <= 0 {
		batch = 20
	}
	return &Runner{repo: repo, jobs: jobs, events: events, interval: interval, batch: batch, logger: logger}
}

func (r *Runner) Run(ctx context.Context) {
	r.wg.Add(2)
	go func() { defer r.wg.Done(); r.runJobs(ctx) }()
	go func() { defer r.wg.Done(); r.runOutbox(ctx) }()
}

func (r *Runner) Wait() { r.wg.Wait() }

func (r *Runner) runJobs(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		if err := r.processJobs(ctx); err != nil && !errors.Is(err, context.Canceled) {
			r.logger.Error("job cycle failed", "error", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (r *Runner) runOutbox(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		if err := r.processOutbox(ctx); err != nil && !errors.Is(err, context.Canceled) {
			r.logger.Error("outbox cycle failed", "error", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
