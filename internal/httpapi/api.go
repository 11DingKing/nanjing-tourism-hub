package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"

	"github.com/11DingKing/nanjing-tourism-hub/internal/service"
)

type API struct {
	service *service.Coordinator
	logger  *slog.Logger
}

func New(coordinator *service.Coordinator, logger *slog.Logger) *API {
	if logger == nil {
		logger = slog.Default()
	}
	return &API{service: coordinator, logger: logger}
}

func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", a.health)
	mux.HandleFunc("GET /readyz", a.ready)
	mux.HandleFunc("POST /v1/login", a.login)
	mux.Handle("POST /v1/logout", a.authenticated(http.HandlerFunc(a.logout)))
	mux.Handle("POST /v1/warnings", a.authenticated(http.HandlerFunc(a.publishWarning)))
	mux.Handle("POST /v1/incidents", a.authenticated(http.HandlerFunc(a.activateIncident)))
	mux.Handle("POST /v1/incidents/{incident}/responses/{region}/escalate", a.authenticated(http.HandlerFunc(a.escalateResponse)))
	mux.Handle("GET /v1/tasks", a.authenticated(http.HandlerFunc(a.listTasks)))
	mux.Handle("POST /v1/tasks", a.authenticated(http.HandlerFunc(a.createTask)))
	mux.Handle("POST /v1/shelters/{shelter}/reservations", a.authenticated(http.HandlerFunc(a.reserveShelter)))
	mux.Handle("POST /v1/teams/{team}/dispatches", a.authenticated(http.HandlerFunc(a.dispatchTeam)))
	mux.Handle("POST /v1/supply-lots/{lot}/allocations", a.authenticated(http.HandlerFunc(a.allocateSupply)))
	mux.Handle("POST /v1/tasks/{task}/receipts", a.authenticated(http.HandlerFunc(a.submitReceipt)))
	mux.Handle("POST /v1/incidents/{incident}/close", a.authenticated(http.HandlerFunc(a.closeIncident)))
	return a.middleware(mux)
}

func (a *API) newRequestID() string {
	var raw [8]byte
	_, _ = rand.Read(raw[:])
	return "req_" + hex.EncodeToString(raw[:])
}
func (a *API) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "alive"})
}
func (a *API) ready(w http.ResponseWriter, r *http.Request) {
	if err := a.service.Readiness(r.Context()); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
