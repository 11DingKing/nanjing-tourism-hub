package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/11DingKing/autumn-grain-resilience/internal/config"
	"github.com/11DingKing/autumn-grain-resilience/internal/httpapi"
	"github.com/11DingKing/autumn-grain-resilience/internal/service"
	"github.com/11DingKing/autumn-grain-resilience/internal/storage"
	"github.com/11DingKing/autumn-grain-resilience/internal/worker"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	db, err := storage.Open(cfg.DatabaseURL)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	if err := storage.Migrate(context.Background(), db); err != nil {
		logger.Error("migrate", "error", err)
		os.Exit(1)
	}
	services := service.NewRegistry(db, logger, cfg)
	w := worker.New(services.Outbox, logger, cfg.WorkerInterval)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go w.Run(ctx)
	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: httpapi.New(services, logger), ReadHeaderTimeout: 5 * time.Second}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server", "error", err)
		}
	}()
	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdown)
	w.Stop()
}
