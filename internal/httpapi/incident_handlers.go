package httpapi

import (
	"encoding/json"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
	"net/http"
	"strconv"
)

func (a *API) activateIncident(w http.ResponseWriter, r *http.Request) {
	actor, _ := actorFrom(r.Context())
	var input struct {
		ID        string               `json:"id"`
		WarningID string               `json:"warning_id"`
		Name      string               `json:"name"`
		Regions   []string             `json:"regions"`
		Level     domain.ResponseLevel `json:"level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Code: "invalid_json", Message: "request body is invalid", RequestID: requestIDFrom(r.Context())})
		return
	}
	value, err := a.service.ActivateIncident(r.Context(), domain.ActivateIncidentCommand{IncidentID: input.ID, WarningID: input.WarningID, Name: input.Name, Regions: input.Regions, Level: input.Level, Actor: actor})
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, value)
}
func (a *API) escalateResponse(w http.ResponseWriter, r *http.Request) {
	actor, _ := actorFrom(r.Context())
	var input struct {
		Version int64                `json:"version"`
		Target  domain.ResponseLevel `json:"target"`
		Reason  string               `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Code: "invalid_json", Message: "request body is invalid", RequestID: requestIDFrom(r.Context())})
		return
	}
	value, err := a.service.EscalateResponse(r.Context(), domain.EscalateResponseCommand{IncidentID: r.PathValue("incident"), RegionID: r.PathValue("region"), FromVersion: input.Version, Target: input.Target, Reason: input.Reason, Actor: actor})
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, value)
}
func (a *API) closeIncident(w http.ResponseWriter, r *http.Request) {
	actor, _ := actorFrom(r.Context())
	version, err := strconv.ParseInt(r.URL.Query().Get("version"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Code: "invalid_version", Message: "version is required", RequestID: requestIDFrom(r.Context())})
		return
	}
	value, err := a.service.CloseIncident(r.Context(), domain.CloseIncidentCommand{IncidentID: r.PathValue("incident"), ExpectedVersion: version, Actor: actor})
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, value)
}
