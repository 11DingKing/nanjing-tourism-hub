package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/11DingKing/autumn-grain-resilience/internal/domain"
	"github.com/11DingKing/autumn-grain-resilience/internal/service"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type API struct {
	Services *service.Registry
	Logger   *slog.Logger
}

func New(s *service.Registry, l *slog.Logger) http.Handler {
	a := &API{Services: s, Logger: l}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", a.health)
	mux.HandleFunc("GET /readyz", a.ready)
	mux.HandleFunc("POST /v1/users", a.register)
	mux.HandleFunc("POST /v1/sessions", a.login)
	mux.HandleFunc("DELETE /v1/sessions/{id}", a.logout)
	mux.HandleFunc("POST /v1/observations", a.observation)
	mux.HandleFunc("POST /v1/field-tasks", a.fieldTask)
	mux.HandleFunc("POST /v1/field-tasks/{id}/claim", a.claim)
	mux.HandleFunc("POST /v1/disasters", a.disaster)
	mux.HandleFunc("POST /v1/drying-reservations", a.drying)
	mux.HandleFunc("POST /v1/aid-cases", a.aid)
	return recoverer(requestID(mux), l)
}
func recoverer(next http.Handler, l *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				l.Error("panic", "error", v)
				writeError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = "req-" + time.Now().UTC().Format("20060102150405.000000000")
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(withRequestID(r.Context(), id)))
	})
}

type ctxKey string

const requestKey ctxKey = "request-id"

func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestKey, id)
}
func reqID(r *http.Request) string { v, _ := r.Context().Value(requestKey).(string); return v }
func (a *API) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"status": "ok"})
}
func (a *API) ready(w http.ResponseWriter, r *http.Request) {
	if err := a.Services.Outbox.DB.QueryRowContext(r.Context(), "SELECT 1").Scan(new(int)); err != nil {
		writeError(w, r, 503, "not_ready", "database unavailable")
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ready"})
}
func decode(r *http.Request, v any) error {
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	return d.Decode(v)
}
func (a *API) register(w http.ResponseWriter, r *http.Request) {
	var in struct{ Name, Role, Password string }
	if err := decode(r, &in); err != nil {
		writeError(w, r, 400, "invalid_json", err.Error())
		return
	}
	u, e := a.Services.Auth.Register(r.Context(), in.Name, in.Role, in.Password)
	if e != nil {
		writeDomainError(w, r, e)
		return
	}
	writeJSON(w, 201, map[string]any{"id": u.ID, "name": u.Name, "role": u.Role})
}
func (a *API) login(w http.ResponseWriter, r *http.Request) {
	var in struct{ UserID, Password string }
	if err := decode(r, &in); err != nil {
		writeError(w, r, 400, "invalid_json", err.Error())
		return
	}
	s, e := a.Services.Auth.Login(r.Context(), in.UserID, in.Password)
	if e != nil {
		writeDomainError(w, r, e)
		return
	}
	writeJSON(w, 201, s)
}
func (a *API) logout(w http.ResponseWriter, r *http.Request) {
	if e := a.Services.Auth.Logout(r.Context(), r.PathValue("id")); e != nil {
		writeDomainError(w, r, e)
		return
	}
	w.WriteHeader(204)
}
func (a *API) observation(w http.ResponseWriter, r *http.Request) {
	var in domain.Observation
	if err := decode(r, &in); err != nil {
		writeError(w, r, 400, "invalid_json", err.Error())
		return
	}
	in.ObservedAt = time.Now().UTC()
	alert, e := a.Services.Monitoring.Record(r.Context(), in, r.Header.Get("X-Region-ID"), reqID(r))
	if e != nil {
		writeDomainError(w, r, e)
		return
	}
	writeJSON(w, 201, alert)
}
func (a *API) fieldTask(w http.ResponseWriter, r *http.Request) {
	var in domain.FieldTask
	if err := decode(r, &in); err != nil {
		writeError(w, r, 400, "invalid_json", err.Error())
		return
	}
	if e := a.Services.FieldTasks.Create(r.Context(), in, r.Header.Get("X-User-ID")); e != nil {
		writeDomainError(w, r, e)
		return
	}
	writeJSON(w, 201, in)
}
func (a *API) claim(w http.ResponseWriter, r *http.Request) {
	ok, e := a.Services.FieldTasks.Claim(r.Context(), r.PathValue("id"), r.Header.Get("X-User-ID"), reqID(r))
	if e != nil {
		writeDomainError(w, r, e)
		return
	}
	if !ok {
		writeError(w, r, 409, "claim_conflict", "task already claimed")
		return
	}
	writeJSON(w, 200, map[string]bool{"claimed": true})
}
func (a *API) disaster(w http.ResponseWriter, r *http.Request) {
	var in domain.DisasterReport
	if err := decode(r, &in); err != nil {
		writeError(w, r, 400, "invalid_json", err.Error())
		return
	}
	c, e := a.Services.Disasters.Report(r.Context(), in, reqID(r))
	if e != nil {
		writeDomainError(w, r, e)
		return
	}
	writeJSON(w, 201, c)
}
func (a *API) drying(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Reservation domain.DryingReservation
		Capacity    float64
	}
	if err := decode(r, &in); err != nil {
		writeError(w, r, 400, "invalid_json", err.Error())
		return
	}
	e := a.Services.Drying.Reserve(r.Context(), in.Reservation, in.Capacity, reqID(r))
	if e != nil {
		writeDomainError(w, r, e)
		return
	}
	writeJSON(w, 201, in.Reservation)
}
func (a *API) aid(w http.ResponseWriter, r *http.Request) {
	var in domain.AidCase
	if err := decode(r, &in); err != nil {
		writeError(w, r, 400, "invalid_json", err.Error())
		return
	}
	e := a.Services.Aid.Create(r.Context(), in, reqID(r))
	if e != nil {
		writeDomainError(w, r, e)
		return
	}
	writeJSON(w, 201, in)
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}, "request_id": reqID(r)})
}
func writeDomainError(w http.ResponseWriter, r *http.Request, e error) {
	status, code := 500, "internal_error"
	switch {
	case errors.Is(e, domain.ErrNotFound):
		status, code = 404, "not_found"
	case errors.Is(e, domain.ErrForbidden):
		status, code = 403, "forbidden"
	case errors.Is(e, domain.ErrConflict):
		status, code = 409, "conflict"
	case errors.Is(e, domain.ErrInvalidState):
		status, code = 409, "invalid_state"
	case errors.Is(e, domain.ErrCapacity):
		status, code = 409, "capacity_unavailable"
	case errors.Is(e, domain.ErrExpired):
		status, code = 410, "expired"
	}
	writeError(w, r, status, code, e.Error())
}

var _ = strings.TrimSpace
