package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
	"github.com/11DingKing/nanjing-tourism-hub/internal/service"
	"github.com/11DingKing/nanjing-tourism-hub/internal/storage/sqlite"
)

type httpFixture struct {
	handler http.Handler
	service *service.Coordinator
	now     time.Time
}

func newHTTPFixture(t *testing.T) httpFixture {
	t.Helper()
	repo, err := sqlite.Open(filepath.Join(t.TempDir(), "http.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	if err := repo.Migrate(context.Background()); err != nil {
		t.Fatalf("migrate database: %v", err)
	}
	now := time.Date(2026, 8, 23, 4, 0, 0, 0, time.UTC)
	coordinator := service.New(repo, func() time.Time { return now }, 8*time.Hour)
	if err := coordinator.Bootstrap(context.Background(), service.BootstrapInput{
		Region: domain.Region{
			ID:        "http-region",
			Name:      "HTTP 测试区域",
			TimeZone:  "Asia/Shanghai",
			Coastal:   true,
			RiskLevel: 4,
			Version:   1,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Users: []service.BootstrapUser{
			{ID: "http-duty", Username: "http-duty", Password: "http-secret", Role: domain.RoleDutyOfficer, RegionID: "http-region"},
		},
	}); err != nil {
		t.Fatalf("bootstrap database: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return httpFixture{handler: New(coordinator, logger).Handler(), service: coordinator, now: now}
}

func TestHealthAndReadinessContracts(t *testing.T) {
	fixture := newHTTPFixture(t)
	for _, path := range []string{"/healthz", "/readyz"} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, path, nil)
			request.Header.Set("X-Request-ID", "health-request")
			response := httptest.NewRecorder()
			fixture.handler.ServeHTTP(response, request)
			if response.Code != http.StatusOK {
				t.Fatalf("got status %d body %s", response.Code, response.Body.String())
			}
			if response.Header().Get("X-Request-ID") != "health-request" {
				t.Fatalf("request id was not echoed: %#v", response.Header())
			}
			var body map[string]string
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body["status"] == "" {
				t.Fatalf("missing status in %#v", body)
			}
		})
	}
}

func TestLoginRejectsBadJSONAndBadCredentials(t *testing.T) {
	fixture := newHTTPFixture(t)
	tests := []struct {
		name   string
		body   string
		status int
		code   string
	}{
		{name: "bad json", body: "{", status: http.StatusBadRequest, code: "invalid_json"},
		{name: "bad password", body: "{\"username\":\"http-duty\",\"password\":\"wrong\"}", status: http.StatusUnauthorized, code: "unauthorized"},
		{name: "unknown user", body: "{\"username\":\"missing\",\"password\":\"wrong\"}", status: http.StatusUnauthorized, code: "unauthorized"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/v1/login", bytes.NewBufferString(test.body))
			response := httptest.NewRecorder()
			fixture.handler.ServeHTTP(response, request)
			if response.Code != test.status {
				t.Fatalf("got status %d want %d body %s", response.Code, test.status, response.Body.String())
			}
			var body errorBody
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if body.Code != test.code {
				t.Fatalf("got code %q want %q", body.Code, test.code)
			}
			if body.RequestID == "" {
				t.Fatal("error response omitted request id")
			}
		})
	}
}

func TestAuthenticatedWarningPublicationThroughHTTP(t *testing.T) {
	fixture := newHTTPFixture(t)
	loginRequest := httptest.NewRequest(http.MethodPost, "/v1/login", bytes.NewBufferString("{\"username\":\"http-duty\",\"password\":\"http-secret\"}"))
	loginResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(loginResponse, loginRequest)
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("login failed: %d %s", loginResponse.Code, loginResponse.Body.String())
	}
	var login service.LoginResult
	if err := json.Unmarshal(loginResponse.Body.Bytes(), &login); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	warning := domain.Warning{
		ID:             "http-warning",
		TyphoonName:    "紫檀",
		Number:         "19",
		Level:          domain.WarningBlue,
		IssuedAt:       fixture.now,
		EffectiveFrom:  fixture.now,
		EffectiveUntil: fixture.now.Add(72 * time.Hour),
		RainfallMM:     380,
		CenterArea:     "北部湾",
		Movement:       "回旋后向东北移动",
	}
	body, err := json.Marshal(warning)
	if err != nil {
		t.Fatalf("encode warning: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/warnings", bytes.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+login.Token)
	request.Header.Set("X-Request-ID", "publish-request")
	response := httptest.NewRecorder()
	fixture.handler.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("publish failed: %d %s", response.Code, response.Body.String())
	}
	var published domain.Warning
	if err := json.Unmarshal(response.Body.Bytes(), &published); err != nil {
		t.Fatalf("decode warning: %v", err)
	}
	if published.ID != warning.ID || published.Status != "published" || published.Version != 1 {
		t.Fatalf("unexpected warning %#v", published)
	}
}

func TestProtectedEndpointRequiresBearerSession(t *testing.T) {
	fixture := newHTTPFixture(t)
	request := httptest.NewRequest(http.MethodGet, "/v1/tasks", nil)
	response := httptest.NewRecorder()
	fixture.handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d body %s", response.Code, response.Body.String())
	}
	var body errorBody
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if body.Code != "unauthorized" || body.Message == "" || body.RequestID == "" {
		t.Fatalf("unexpected error contract %#v", body)
	}
}
