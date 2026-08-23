package httpapi

import (
	"encoding/json"
	"net/http"
)

func (a *API) login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Code: "invalid_json", Message: "request body is invalid", RequestID: requestIDFrom(r.Context())})
		return
	}
	result, err := a.service.Login(r.Context(), input.Username, input.Password)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *API) logout(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFrom(r.Context())
	if !ok {
		writeError(w, r, http.ErrNoCookie)
		return
	}
	var input struct {
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Code: "invalid_json", Message: "request body is invalid", RequestID: requestIDFrom(r.Context())})
		return
	}
	if err := a.service.Logout(r.Context(), input.SessionID, actor); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
