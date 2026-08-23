package httpapi

import (
	"encoding/json"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
	"net/http"
)

func (a *API) reserveShelter(w http.ResponseWriter, r *http.Request) {
	actor, _ := actorFrom(r.Context())
	var cmd domain.ReserveShelterCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Code: "invalid_json", Message: "request body is invalid", RequestID: requestIDFrom(r.Context())})
		return
	}
	cmd.ShelterID = r.PathValue("shelter")
	cmd.Actor = actor
	value, err := a.service.ReserveShelter(r.Context(), cmd)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, value)
}
func (a *API) dispatchTeam(w http.ResponseWriter, r *http.Request) {
	actor, _ := actorFrom(r.Context())
	var cmd domain.DispatchTeamCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Code: "invalid_json", Message: "request body is invalid", RequestID: requestIDFrom(r.Context())})
		return
	}
	cmd.TeamID = r.PathValue("team")
	cmd.Actor = actor
	value, err := a.service.DispatchTeam(r.Context(), cmd)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, value)
}
func (a *API) allocateSupply(w http.ResponseWriter, r *http.Request) {
	actor, _ := actorFrom(r.Context())
	var cmd domain.AllocateSupplyCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Code: "invalid_json", Message: "request body is invalid", RequestID: requestIDFrom(r.Context())})
		return
	}
	cmd.LotID = r.PathValue("lot")
	cmd.Actor = actor
	value, err := a.service.AllocateSupply(r.Context(), cmd)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, value)
}
