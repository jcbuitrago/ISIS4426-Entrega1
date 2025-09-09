package routers

import (
	"context"
	"encoding/json"
	"net/http"

	"ISIS4426-Entrega1/app/async"
	"github.com/gorilla/mux"
)

type JobsHandler struct{ enq *async.Enqueuer }

func NewJobsHandler(e *async.Enqueuer) *JobsHandler { return &JobsHandler{enq: e} }

// GET /api/jobs/{id}
func (h *JobsHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		http.Error(w, "id faltante", http.StatusBadRequest)
		return
	}
	status, err := h.enq.GetStatus(context.Background(), id)
	if err != nil {
		http.Error(w, "no encontrado", http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"status": status})
}
