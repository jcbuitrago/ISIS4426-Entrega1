package routers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type StatusGetter interface {
	GetStatus(ctx context.Context, jobID string) (string, error)
}

type JobsHandler struct {
	enqueuer StatusGetter
}

func NewJobsHandler(e StatusGetter) *JobsHandler {
	return &JobsHandler{enqueuer: e}
}

// GET /api/jobs/{id}
func (h *JobsHandler) GetJobStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]

	status, err := h.enqueuer.GetStatus(r.Context(), jobID)
	if err != nil {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"job_id": jobID,
		"status": status,
	})
}
