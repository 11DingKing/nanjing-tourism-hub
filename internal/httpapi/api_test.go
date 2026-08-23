package httpapi

import (
	"context"
	"github.com/11DingKing/autumn-grain-resilience/internal/config"
	"github.com/11DingKing/autumn-grain-resilience/internal/service"
	"github.com/11DingKing/autumn-grain-resilience/internal/storage"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthAndRequestID(t *testing.T) {
	db, e := storage.Open(":memory:")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	if e = storage.Migrate(context.Background(), db); e != nil {
		t.Fatal(e)
	}
	h := New(service.NewRegistry(db, slog.Default(), config.Config{}), slog.Default())
	r := httptest.NewRequest("GET", "/healthz", nil)
	r.Header.Set("X-Request-ID", "req-test")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 200 || w.Header().Get("X-Request-ID") != "req-test" {
		t.Fatalf("code=%d id=%s", w.Code, w.Header().Get("X-Request-ID"))
	}
	b, _ := io.ReadAll(w.Result().Body)
	if !strings.Contains(string(b), "ok") {
		t.Fatal(string(b))
	}
}
