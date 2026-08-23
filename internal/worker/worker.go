package worker

import (
	"context"
	"github.com/11DingKing/autumn-grain-resilience/internal/repository"
	"log/slog"
	"time"
)

type Worker struct {
	Outbox   *repository.OutboxRepo
	Logger   *slog.Logger
	Interval time.Duration
	stop     chan struct{}
}

func New(out *repository.OutboxRepo, l *slog.Logger, interval time.Duration) *Worker {
	return &Worker{Outbox: out, Logger: l, Interval: interval, stop: make(chan struct{})}
}
func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stop:
			return
		case now := <-ticker.C:
			w.process(ctx, now)
		}
	}
}
func (w *Worker) Stop() {
	select {
	case <-w.stop:
	default:
		close(w.stop)
	}
}
func (w *Worker) process(ctx context.Context, now time.Time) {
	events, e := w.Outbox.Ready(ctx, now, 20)
	if e != nil {
		w.Logger.Error("outbox read", "error", e)
		return
	}
	for _, event := range events {
		w.Logger.Info("dispatch event", "topic", event.Topic, "aggregate", event.AggregateID)
		if e = w.Outbox.Mark(ctx, event.ID, "sent", event.Attempts, now); e != nil {
			w.Logger.Error("outbox mark", "error", e)
		}
	}
}
