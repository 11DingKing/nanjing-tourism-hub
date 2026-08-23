package worker

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
	"github.com/11DingKing/nanjing-tourism-hub/internal/repository"
	"github.com/11DingKing/nanjing-tourism-hub/internal/storage/sqlite"
)

type recordingHandlers struct {
	mu         sync.Mutex
	jobs       []string
	events     []string
	jobError   error
	eventError error
}

func (h *recordingHandlers) HandleJob(ctx context.Context, kind, aggregateID string, payload []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.jobs = append(h.jobs, kind+":"+aggregateID+":"+string(payload))
	return h.jobError
}

func (h *recordingHandlers) Publish(ctx context.Context, topic, aggregateID string, payload []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.events = append(h.events, topic+":"+aggregateID+":"+string(payload))
	return h.eventError
}

func newWorkerStore(t *testing.T) *sqlite.Store {
	t.Helper()
	store, err := sqlite.Open(filepath.Join(t.TempDir(), "worker.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	if err := store.Migrate(context.Background()); err != nil {
		t.Fatalf("migrate database: %v", err)
	}
	return store
}

func seedWorkerRecords(t *testing.T, store *sqlite.Store, now time.Time) {
	t.Helper()
	err := store.WithinTx(context.Background(), func(tx repository.Tx) error {
		if err := tx.InsertJob(context.Background(), domain.Job{
			ID:          "job-1",
			Kind:        "reservation.expire",
			AggregateID: "reservation-1",
			Payload:     []byte("{\"reservation_id\":\"reservation-1\"}"),
			Status:      "pending",
			MaxAttempts: 3,
			NextRunAt:   now,
			CreatedAt:   now,
			UpdatedAt:   now,
		}); err != nil {
			return err
		}
		return tx.InsertOutbox(context.Background(), domain.OutboxEvent{
			ID:            "event-1",
			Topic:         "warning.published",
			AggregateType: "warning",
			AggregateID:   "warning-1",
			Payload:       []byte("{\"warning_id\":\"warning-1\"}"),
			Status:        "pending",
			NextAttemptAt: now,
			CreatedAt:     now,
			UpdatedAt:     now,
		})
	})
	if err != nil {
		t.Fatalf("seed records: %v", err)
	}
}

func TestWorkerPersistsCompletionAfterHandlerSuccess(t *testing.T) {
	store := newWorkerStore(t)
	now := time.Now().UTC().Add(-time.Minute)
	seedWorkerRecords(t, store, now)
	handlers := &recordingHandlers{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	runner := New(store, handlers, handlers, time.Second, 10, logger)
	if err := runner.processJobs(context.Background()); err != nil {
		t.Fatalf("process jobs: %v", err)
	}
	if err := runner.processOutbox(context.Background()); err != nil {
		t.Fatalf("process outbox: %v", err)
	}
	handlers.mu.Lock()
	defer handlers.mu.Unlock()
	if len(handlers.jobs) != 1 {
		t.Fatalf("got jobs %#v", handlers.jobs)
	}
	if len(handlers.events) != 1 {
		t.Fatalf("got events %#v", handlers.events)
	}
	jobs, err := store.ClaimDueJobs(context.Background(), time.Now().UTC().Add(time.Hour), 10, time.Minute)
	if err != nil {
		t.Fatalf("claim completed jobs: %v", err)
	}
	if len(jobs) != 0 {
		t.Fatalf("completed job remained claimable: %#v", jobs)
	}
	events, err := store.ClaimOutbox(context.Background(), time.Now().UTC().Add(time.Hour), 10)
	if err != nil {
		t.Fatalf("claim completed events: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("published event remained claimable: %#v", events)
	}
}

func TestWorkerSchedulesRetryAfterHandlerFailure(t *testing.T) {
	store := newWorkerStore(t)
	now := time.Now().UTC().Add(-time.Minute)
	seedWorkerRecords(t, store, now)
	handlers := &recordingHandlers{jobError: errors.New("temporary database outage"), eventError: errors.New("temporary broker outage")}
	runner := New(store, handlers, handlers, time.Second, 10, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err := runner.processJobs(context.Background()); err != nil {
		t.Fatalf("process failed job: %v", err)
	}
	if err := runner.processOutbox(context.Background()); err != nil {
		t.Fatalf("process failed event: %v", err)
	}
	jobs, err := store.ClaimDueJobs(context.Background(), time.Now().UTC().Add(time.Hour), 10, time.Minute)
	if err != nil {
		t.Fatalf("claim retry jobs: %v", err)
	}
	if len(jobs) != 1 || jobs[0].Attempts != 1 || jobs[0].LastError == "" {
		t.Fatalf("job retry state not persisted: %#v", jobs)
	}
	events, err := store.ClaimOutbox(context.Background(), time.Now().UTC().Add(time.Hour), 10)
	if err != nil {
		t.Fatalf("claim retry events: %v", err)
	}
	if len(events) != 1 || events[0].Attempts != 1 || events[0].LastError == "" {
		t.Fatalf("event retry state not persisted: %#v", events)
	}
}

func TestWorkerStopsOnContextCancellation(t *testing.T) {
	store := newWorkerStore(t)
	handlers := &recordingHandlers{}
	runner := New(store, handlers, handlers, 5*time.Millisecond, 5, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ctx, cancel := context.WithCancel(context.Background())
	runner.Run(ctx)
	cancel()
	done := make(chan struct{})
	go func() { runner.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop after context cancellation")
	}
}
