package httpapi

import (
	"encoding/json"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
	"net/http"
)

func (a *API) publishWarning(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFrom(r.Context())
	if !ok {
		writeError(w, r, domain.ErrUnauthorized)
		return
	}
	var warning domain.Warning
	if err := json.NewDecoder(r.Body).Decode(&warning); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Code: "invalid_json", Message: "request body is invalid", RequestID: requestIDFrom(r.Context())})
		return
	}
	value, err := a.service.PublishWarning(r.Context(), domain.PublishWarningCommand{Warning: warning, Actor: actor})
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, value)
}
