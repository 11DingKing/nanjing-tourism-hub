package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/11DingKing/nanjing-tourism-hub/internal/config"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
	"github.com/11DingKing/nanjing-tourism-hub/internal/httpapi"
	"github.com/11DingKing/nanjing-tourism-hub/internal/service"
	"github.com/11DingKing/nanjing-tourism-hub/internal/storage/sqlite"
	"github.com/11DingKing/nanjing-tourism-hub/internal/worker"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	repo, err := sqlite.Open(cfg.DatabasePath)
	if err != nil {
		return err
	}
	defer repo.Close()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := repo.Migrate(ctx); err != nil {
		return err
	}
	coordinator := service.New(repo, time.Now, cfg.SessionTTL)
	if err := coordinator.Bootstrap(ctx, defaultBootstrap(time.Now().UTC())); err != nil {
		return err
	}
	handlers := worker.LocalHandlers{Logger: logger}
	runner := worker.New(repo, handlers, handlers, cfg.WorkerInterval, cfg.WorkerBatch, logger)
	runner.Run(ctx)
	server := &http.Server{Addr: cfg.HTTPAddr, Handler: httpapi.New(coordinator, logger).Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("server listening", "address", cfg.HTTPAddr)
		serverErrors <- server.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return err
	}
	stop()
	runner.Wait()
	return nil
}

func defaultBootstrap(now time.Time) service.BootstrapInput {
	region := domain.Region{ID: "beibuwan", Name: "北部湾应急协调区", TimeZone: "Asia/Shanghai", Coastal: true, RiskLevel: 4, Version: 1, CreatedAt: now, UpdatedAt: now}
	return service.BootstrapInput{Region: region, Users: []service.BootstrapUser{{ID: "duty-1", Username: "duty", Password: "change-me-duty", Role: domain.RoleDutyOfficer, RegionID: region.ID}, {ID: "dispatch-1", Username: "dispatcher", Password: "change-me-dispatch", Role: domain.RoleDispatcher, RegionID: region.ID}, {ID: "review-1", Username: "reviewer", Password: "change-me-review", Role: domain.RoleReviewer, RegionID: region.ID}, {ID: "field-1", Username: "field", Password: "change-me-field", Role: domain.RoleFieldLead, RegionID: region.ID}}, Shelters: []domain.Shelter{{ID: "shelter-hk-1", RegionID: region.ID, Name: "海口东部避险点", Capacity: 800, Status: domain.ResourceAvailable}}, Teams: []domain.Team{{ID: "team-drainage-1", RegionID: region.ID, Name: "排涝一队", Specialty: "urban_drainage", Status: domain.ResourceAvailable, LeaderID: "field-1"}, {ID: "team-rescue-1", RegionID: region.ID, Name: "抢险一队", Specialty: "rescue", Status: domain.ResourceAvailable, LeaderID: "field-1"}}, SupplyLots: []domain.SupplyLot{{ID: "lot-water-1", RegionID: region.ID, Kind: "drinking_water", Quantity: 5000, ExpiresAt: now.Add(90 * 24 * time.Hour), Status: domain.ResourceAvailable}}}
}
