package routers

import (
	"net/http"

	"ISIS4426-Entrega1/app/async"

	"github.com/gorilla/mux"
)

type JobsHandler struct{ enq *async.Enqueuer }

func NewJobsHandler(enq *async.Enqueuer) *JobsHandler { return &JobsHandler{enq: enq} }

func (h *JobsHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	st, err := h.enq.GetStatus(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"job_id":"` + id + `","status":"` + st + `"}`))
}
