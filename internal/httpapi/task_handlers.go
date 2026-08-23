package httpapi

import (
	"encoding/json"
	"github.com/11DingKing/nanjing-tourism-hub/internal/domain"
	"net/http"
	"strconv"
	"time"
)

func (a *API) listTasks(w http.ResponseWriter, r *http.Request) {
	actor, _ := actorFrom(r.Context())
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	filter := domain.TaskFilter{IncidentID: r.URL.Query().Get("incident_id"), RegionID: r.URL.Query().Get("region_id"), Status: domain.TaskStatus(r.URL.Query().Get("status")), Kind: r.URL.Query().Get("kind"), Page: domain.Page{Limit: limit, Offset: offset, Sort: r.URL.Query().Get("sort"), Direction: r.URL.Query().Get("direction")}}
	page, err := a.service.ListTasks(r.Context(), filter, actor)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}
func (a *API) createTask(w http.ResponseWriter, r *http.Request) {
	actor, _ := actorFrom(r.Context())
	var task domain.DistrictTask
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Code: "invalid_json", Message: "request body is invalid", RequestID: requestIDFrom(r.Context())})
		return
	}
	value, err := a.service.CreateTask(r.Context(), domain.CreateTaskCommand{Task: task, Actor: actor})
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, value)
}
func (a *API) submitReceipt(w http.ResponseWriter, r *http.Request) {
	actor, _ := actorFrom(r.Context())
	var input struct {
		Receipt     domain.Receipt `json:"receipt"`
		TaskVersion int64          `json:"task_version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Code: "invalid_json", Message: "request body is invalid", RequestID: requestIDFrom(r.Context())})
		return
	}
	input.Receipt.TaskID = r.PathValue("task")
	value, err := a.service.SubmitReceipt(r.Context(), domain.SubmitReceiptCommand{Receipt: input.Receipt, TaskVersion: input.TaskVersion, Actor: actor})
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, value)
}

var _ = time.RFC3339
