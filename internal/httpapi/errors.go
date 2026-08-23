package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
	"net/http"
)

type errorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusInternalServerError
	code := "internal_error"
	message := "request could not be completed"
	switch {
	case errors.Is(err, domain.ErrUnauthorized):
		status = http.StatusUnauthorized
		code = "unauthorized"
		message = "authentication is required"
	case errors.Is(err, domain.ErrForbidden):
		status = http.StatusForbidden
		code = "forbidden"
		message = "operation is not allowed"
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
		code = "not_found"
		message = "resource was not found"
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrInvalidState), errors.Is(err, domain.ErrCapacity):
		status = http.StatusConflict
		code = "conflict"
		message = err.Error()
	case errors.Is(err, domain.ErrExpired):
		status = http.StatusGone
		code = "expired"
		message = err.Error()
	case errors.Is(err, domain.ErrCanceled), errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		status = http.StatusRequestTimeout
		code = "canceled"
		message = "request was canceled"
	}
	writeJSON(w, status, errorBody{Code: code, Message: message, RequestID: requestIDFrom(r.Context())})
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
