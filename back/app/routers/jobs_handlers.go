package routers

import (
	"context"
	"encoding/json"
	"log"
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
		log.Printf("[api] jobs: not found job_id=%s", jobID)
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}
	log.Printf("[api] jobs: status job_id=%s status=%s", jobID, status)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"job_id": jobID,
		"status": status,
	})
}
