package httpapi

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
)

func (a *API) middleware(next http.Handler) http.Handler {
	return a.recoverer(a.requestID(a.logging(next)))
}

func (a *API) requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if id == "" {
			id = a.newRequestID()
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(withRequestID(r.Context(), id)))
	})
}

func (a *API) logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		a.logger.Info("http request", "method", r.Method, "path", r.URL.Path, "request_id", requestIDFrom(r.Context()), "duration_ms", time.Since(start).Milliseconds())
	})
}

func (a *API) recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if value := recover(); value != nil {
				a.logger.Error("http panic", slog.Any("panic", value), "request_id", requestIDFrom(r.Context()))
				writeError(w, r, domain.Wrap(domain.ErrDependency, "http", "request", requestIDFrom(r.Context()), "panic recovered", nil))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (a *API) authenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := strings.TrimSpace(r.Header.Get("Authorization"))
		if !strings.HasPrefix(header, "Bearer ") {
			writeError(w, r, domain.ErrUnauthorized)
			return
		}
		actor, err := a.service.Authenticate(r.Context(), strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
		if err != nil {
			writeError(w, r, err)
			return
		}
		actor.RequestID = requestIDFrom(r.Context())
		next.ServeHTTP(w, r.WithContext(withActor(r.Context(), actor)))
	})
}
