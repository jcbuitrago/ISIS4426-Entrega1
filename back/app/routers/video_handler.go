package routers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"ISIS4426-Entrega1/app/async"
)

type VideosHandler struct {
	enqueuer *async.Enqueuer
}

func NewVideosHandler(enq *async.Enqueuer) *VideosHandler {
	return &VideosHandler{enqueuer: enq}
}

func (h *VideosHandler) Create(w http.ResponseWriter, r *http.Request) {
	const maxUpload = 200 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "multipart parse error", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	f, hdr, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file required", http.StatusBadRequest)
		return
	}
	defer f.Close()

	base := "/data/uploads" // 👈 volumen compartido
	_ = os.MkdirAll(base, 0o775)
	tmpPath := filepath.Join(base, filepath.Base(hdr.Filename))

	out, err := os.Create(tmpPath)
	if err != nil {
		http.Error(w, "cannot save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, f); err != nil {
		http.Error(w, "cannot write file", http.StatusInternalServerError)
		return
	}

	jobID, err := h.enqueuer.EnqueueVideoProcessing(r.Context(), 1, title, tmpPath)
	if err != nil {
		http.Error(w, "queue error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]string{"job_id": jobID, "status": "queued"})
}
