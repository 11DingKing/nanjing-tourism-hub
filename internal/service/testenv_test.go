package service

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
	"github.com/11DingKing/nanjing-tourism-hub/internal/storage/sqlite"
)

type testEnvironment struct {
	repo       *sqlite.Store
	service    *Coordinator
	now        time.Time
	duty       domain.Actor
	dispatcher domain.Actor
	reviewer   domain.Actor
	field      domain.Actor
}

func newTestEnvironment(t *testing.T) *testEnvironment {
	t.Helper()
	path := filepath.Join(t.TempDir(), "response-hub.db")
	repo, err := sqlite.Open(path)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() {
		if err := repo.Close(); err != nil {
			t.Errorf("close database: %v", err)
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("migrate database: %v", err)
	}
	now := time.Date(2026, 8, 23, 2, 30, 0, 0, time.UTC)
	coordinator := New(repo, func() time.Time { return now }, 8*time.Hour)
	input := BootstrapInput{
		Region: domain.Region{
			ID:        "coast-a",
			Name:      "北部湾沿海协调区",
			TimeZone:  "Asia/Shanghai",
			Coastal:   true,
			RiskLevel: 4,
			Version:   1,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Users: []BootstrapUser{
			{ID: "duty-1", Username: "duty", Password: "duty-secret", Role: domain.RoleDutyOfficer, RegionID: "coast-a"},
			{ID: "dispatch-1", Username: "dispatch", Password: "dispatch-secret", Role: domain.RoleDispatcher, RegionID: "coast-a"},
			{ID: "review-1", Username: "review", Password: "review-secret", Role: domain.RoleReviewer, RegionID: "coast-a"},
			{ID: "field-1", Username: "field", Password: "field-secret", Role: domain.RoleFieldLead, RegionID: "coast-a"},
		},
		Shelters: []domain.Shelter{
			{ID: "shelter-1", RegionID: "coast-a", Name: "沿海一号避险点", Capacity: 100, Status: domain.ResourceAvailable},
			{ID: "shelter-2", RegionID: "coast-a", Name: "内陆备用避险点", Capacity: 300, Status: domain.ResourceAvailable},
		},
		Teams: []domain.Team{
			{ID: "team-1", RegionID: "coast-a", Name: "排涝一队", Specialty: "drainage", Status: domain.ResourceAvailable, LeaderID: "field-1"},
			{ID: "team-2", RegionID: "coast-a", Name: "抢险二队", Specialty: "rescue", Status: domain.ResourceAvailable, LeaderID: "field-1"},
			{ID: "team-3", RegionID: "coast-a", Name: "通信保障队", Specialty: "communications", Status: domain.ResourceAvailable, LeaderID: "field-1"},
		},
		SupplyLots: []domain.SupplyLot{
			{ID: "lot-water", RegionID: "coast-a", Kind: "drinking_water", Quantity: 200, ExpiresAt: now.Add(30 * 24 * time.Hour), Status: domain.ResourceAvailable},
			{ID: "lot-pump", RegionID: "coast-a", Kind: "portable_pump", Quantity: 20, ExpiresAt: now.Add(365 * 24 * time.Hour), Status: domain.ResourceAvailable},
		},
	}
	if err := coordinator.Bootstrap(ctx, input); err != nil {
		t.Fatalf("bootstrap environment: %v", err)
	}
	return &testEnvironment{
		repo:    repo,
		service: coordinator,
		now:     now,
		duty: domain.Actor{
			UserID:    "duty-1",
			Role:      domain.RoleDutyOfficer,
			RegionID:  "coast-a",
			RequestID: "request-duty",
		},
		dispatcher: domain.Actor{
			UserID:    "dispatch-1",
			Role:      domain.RoleDispatcher,
			RegionID:  "coast-a",
			RequestID: "request-dispatch",
		},
		reviewer: domain.Actor{
			UserID:    "review-1",
			Role:      domain.RoleReviewer,
			RegionID:  "coast-a",
			RequestID: "request-review",
		},
		field: domain.Actor{
			UserID:    "field-1",
			Role:      domain.RoleFieldLead,
			RegionID:  "coast-a",
			RequestID: "request-field",
		},
	}
}

func (e *testEnvironment) seedIncident(t *testing.T, incidentID string) domain.Incident {
	t.Helper()
	ctx := context.Background()
	warning, err := e.service.PublishWarning(ctx, domain.PublishWarningCommand{
		Warning: domain.Warning{
			ID:             "warning-" + incidentID,
			TyphoonName:    "紫檀",
			Number:         "19",
			Level:          domain.WarningBlue,
			IssuedAt:       e.now,
			EffectiveFrom:  e.now,
			EffectiveUntil: e.now.Add(72 * time.Hour),
			RainfallMM:     350,
			CenterArea:     "北部湾",
			Movement:       "回旋后向东北移动",
		},
		Actor: e.duty,
	})
	if err != nil {
		t.Fatalf("publish warning: %v", err)
	}
	incident, err := e.service.ActivateIncident(ctx, domain.ActivateIncidentCommand{
		IncidentID: incidentID,
		WarningID:  warning.ID,
		Name:       "紫檀持续强降雨响应",
		Regions:    []string{"coast-a"},
		Level:      domain.ResponseIV,
		Actor:      e.duty,
	})
	if err != nil {
		t.Fatalf("activate incident: %v", err)
	}
	return incident
}

func (e *testEnvironment) seedTask(t *testing.T, incidentID, taskID string) domain.DistrictTask {
	t.Helper()
	task, err := e.service.CreateTask(context.Background(), domain.CreateTaskCommand{
		Task: domain.DistrictTask{
			ID:         taskID,
			IncidentID: incidentID,
			RegionID:   "coast-a",
			Kind:       "drainage",
			Summary:    "排查低洼路段并布置移动泵",
			Priority:   4,
			Deadline:   e.now.Add(12 * time.Hour),
		},
		Actor: e.dispatcher,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	return task
}
