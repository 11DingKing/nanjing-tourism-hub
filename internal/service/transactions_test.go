package service

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
	"github.com/11DingKing/nanjing-tourism-hub/internal/storage/sqlite"
)

func TestActivationRollsBackIncidentWhenRegionSetIsInvalid(t *testing.T) {
	env := newTestEnvironment(t)
	warning, err := env.service.PublishWarning(context.Background(), domain.PublishWarningCommand{
		Warning: domain.Warning{
			ID:             "warning-rollback",
			TyphoonName:    "紫檀",
			Number:         "19",
			Level:          domain.WarningBlue,
			IssuedAt:       env.now,
			EffectiveFrom:  env.now,
			EffectiveUntil: env.now.Add(72 * time.Hour),
			RainfallMM:     420,
			CenterArea:     "北部湾",
			Movement:       "缓慢回旋",
		},
		Actor: env.duty,
	})
	if err != nil {
		t.Fatalf("publish warning: %v", err)
	}
	_, err = env.service.ActivateIncident(context.Background(), domain.ActivateIncidentCommand{
		IncidentID: "incident-rollback",
		WarningID:  warning.ID,
		Name:       "事务回滚检查",
		Regions:    []string{"coast-a", "coast-a"},
		Level:      domain.ResponseIV,
		Actor:      env.duty,
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected duplicate region conflict, got %v", err)
	}
	incident, err := env.service.ActivateIncident(context.Background(), domain.ActivateIncidentCommand{
		IncidentID: "incident-rollback",
		WarningID:  warning.ID,
		Name:       "事务回滚检查",
		Regions:    []string{"coast-a"},
		Level:      domain.ResponseIV,
		Actor:      env.duty,
	})
	if err != nil {
		t.Fatalf("incident insert was not rolled back: %v", err)
	}
	if incident.ID != "incident-rollback" {
		t.Fatalf("unexpected incident %#v", incident)
	}
}

func TestCanceledContextDoesNotBeginBusinessTransaction(t *testing.T) {
	env := newTestEnvironment(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := env.service.PublishWarning(ctx, domain.PublishWarningCommand{
		Warning: domain.Warning{
			ID:             "warning-canceled",
			TyphoonName:    "紫檀",
			Number:         "19",
			Level:          domain.WarningBlue,
			IssuedAt:       env.now,
			EffectiveFrom:  env.now,
			EffectiveUntil: env.now.Add(time.Hour),
			RainfallMM:     200,
			CenterArea:     "北部湾",
			Movement:       "向东北移动",
		},
		Actor: env.duty,
	})
	if err == nil {
		t.Fatal("expected canceled publish to fail")
	}
	if !errors.Is(err, context.Canceled) && !errors.Is(err, domain.ErrCanceled) {
		t.Fatalf("cancellation was not preserved in error chain: %v", err)
	}
}

func TestSessionsAndOperationalStateSurviveDatabaseRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "restart.db")
	now := time.Date(2026, 8, 23, 3, 0, 0, 0, time.UTC)
	repo, err := sqlite.Open(path)
	if err != nil {
		t.Fatalf("open first database: %v", err)
	}
	if err := repo.Migrate(context.Background()); err != nil {
		t.Fatalf("migrate first database: %v", err)
	}
	coordinator := New(repo, func() time.Time { return now }, 8*time.Hour)
	input := BootstrapInput{
		Region: domain.Region{
			ID:        "restart-region",
			Name:      "重启恢复区域",
			TimeZone:  "Asia/Shanghai",
			Coastal:   true,
			RiskLevel: 3,
			Version:   1,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Users: []BootstrapUser{
			{ID: "restart-duty", Username: "restart-duty", Password: "restart-secret", Role: domain.RoleDutyOfficer, RegionID: "restart-region"},
		},
		Shelters: []domain.Shelter{
			{ID: "restart-shelter", RegionID: "restart-region", Name: "恢复避险点", Capacity: 40, Status: domain.ResourceAvailable},
		},
	}
	if err := coordinator.Bootstrap(context.Background(), input); err != nil {
		t.Fatalf("bootstrap first database: %v", err)
	}
	login, err := coordinator.Login(context.Background(), "restart-duty", "restart-secret")
	if err != nil {
		t.Fatalf("login before restart: %v", err)
	}
	if err := repo.Close(); err != nil {
		t.Fatalf("close first database: %v", err)
	}
	reopened, err := sqlite.Open(path)
	if err != nil {
		t.Fatalf("reopen database: %v", err)
	}
	defer reopened.Close()
	if err := reopened.Migrate(context.Background()); err != nil {
		t.Fatalf("repeat migrations after restart: %v", err)
	}
	restored := New(reopened, func() time.Time { return now.Add(time.Minute) }, 8*time.Hour)
	actor, err := restored.Authenticate(context.Background(), login.Token)
	if err != nil {
		t.Fatalf("authenticate restored session: %v", err)
	}
	if actor.UserID != "restart-duty" || actor.Role != domain.RoleDutyOfficer {
		t.Fatalf("unexpected restored actor %#v", actor)
	}
}

func TestLogoutRevokesServerSideSession(t *testing.T) {
	env := newTestEnvironment(t)
	login, err := env.service.Login(context.Background(), "dispatch", "dispatch-secret")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	actor, err := env.service.Authenticate(context.Background(), login.Token)
	if err != nil {
		t.Fatalf("authenticate active session: %v", err)
	}
	if actor.UserID != "dispatch-1" {
		t.Fatalf("unexpected actor %#v", actor)
	}
	if err := env.service.Logout(context.Background(), login.SessionID, domain.Actor{
		UserID:    actor.UserID,
		Role:      actor.Role,
		RegionID:  actor.RegionID,
		RequestID: "logout-request",
	}); err != nil {
		t.Fatalf("logout: %v", err)
	}
	_, err = env.service.Authenticate(context.Background(), login.Token)
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("revoked session remained active: %v", err)
	}
}

func TestExpiredSessionIsRejectedWithoutDeletingHistory(t *testing.T) {
	env := newTestEnvironment(t)
	login, err := env.service.Login(context.Background(), "review", "review-secret")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	env.service.now = func() time.Time { return env.now.Add(9 * time.Hour) }
	_, err = env.service.Authenticate(context.Background(), login.Token)
	if !errors.Is(err, domain.ErrExpired) {
		t.Fatalf("expected expiration error, got %v", err)
	}
}
