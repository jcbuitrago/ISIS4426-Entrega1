package routers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"ISIS4426-Entrega1/app/async"
	"ISIS4426-Entrega1/app/models"
	"ISIS4426-Entrega1/app/repos"
	"ISIS4426-Entrega1/app/services"

	"github.com/gorilla/mux"
)

type VideosHandler struct {
	enqueuer *async.Enqueuer
	svc      *services.VideoService
}

func NewVideosHandler(enq *async.Enqueuer, svc *services.VideoService) *VideosHandler {
	return &VideosHandler{enqueuer: enq, svc: svc}
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

	base := "/data/uploads"
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
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Video subido correctamente. Procesamiento en curso.", "task_id": jobID})
}

func (h *VideosHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	userIDStr := q.Get("user_id")

	var (
		items []models.Video
		err   error
	)
	if userIDStr != "" {
		uid, convErr := strconv.Atoi(userIDStr)
		if convErr != nil {
			http.Error(w, "user_id inválido", http.StatusBadRequest)
			return
		}
		items, err = h.svc.ListByUser(r.Context(), uid, limit, offset)
	} else {
		items, err = h.svc.List(r.Context(), limit, offset)
	}
	if err != nil {
		http.Error(w, "Error al consultar videos", http.StatusInternalServerError)
		return
	}

	type respItem struct {
		VideoID      string `json:"video_id"`
		Title        string `json:"title"`
		Status       string `json:"status"`
		UploadedAt   string `json:"uploaded_at"`
		ProcessedAt  string `json:"processed_at,omitempty"`
		ProcessedURL string `json:"processed_url,omitempty"`
	}

	out := make([]respItem, 0, len(items))
	for _, v := range items {
		it := respItem{
			VideoID:    strconv.Itoa(v.VideoID),
			Title:      v.Title,
			Status:     string(v.Status),
			UploadedAt: v.UploadedAt.UTC().Format(time.RFC3339),
		}
		if !v.ProcessedAt.IsZero() {
			it.ProcessedAt = v.ProcessedAt.UTC().Format(time.RFC3339)
		}
		if v.ProcessedURL != "" {
			it.ProcessedURL = v.ProcessedURL
		}
		out = append(out, it)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (h *VideosHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1) intenta path param: /api/videos/{id}
	idStr := mux.Vars(r)["id"]
	if idStr == "" {
		// 2) fallback: ?id=123
		idStr = r.URL.Query().Get("id")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	v, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		// si tu repo expone ErrNotFound, devuelve 404
		if errors.Is(err, repos.ErrNotFound) {
			http.Error(w, "video no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "video no encontrado", http.StatusNotFound)
		return
	}

	_ = json.NewEncoder(w).Encode(v)
}

func (h *VideosHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	if idStr == "" {
		idStr = r.URL.Query().Get("id")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		println(err)
		http.Error(w, "error al eliminar video", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Video eliminado correctamente.", "video_id": idStr})
}
