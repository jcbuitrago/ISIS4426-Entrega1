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
	"ISIS4426-Entrega1/app/middleware"

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
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok { http.Error(w, "unauthorized", http.StatusUnauthorized); return }

	const maxUpload = 100 << 20 // 100MB
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "multipart parse error", http.StatusBadRequest)
		return
	}
	title := r.FormValue("title")
	f, hdr, err := r.FormFile("video_file")
	if err != nil { http.Error(w, "archivo faltante", http.StatusBadRequest); return }
	defer f.Close()

	// Guardar temporalmente
	base := "/data/uploads"
	_ = os.MkdirAll(base, 0o775)
	tmpPath := filepath.Join(base, filepath.Base(hdr.Filename))
	out, err := os.Create(tmpPath)
	if err != nil { http.Error(w, "cannot save file", http.StatusInternalServerError); return }
	defer out.Close()
	if _, err := io.Copy(out, f); err != nil { http.Error(w, "cannot write file", http.StatusInternalServerError); return }

	// Crear registro en DB con status uploaded
	created, err := h.svc.Create(uid, title, tmpPath)
	if err != nil {
		http.Error(w, "error al crear registro", http.StatusInternalServerError)
		return
	}

	// Encolar trabajo
	jobID, err := h.enqueuer.EnqueueVideoProcessing(r.Context(), created.VideoID, uid, title, tmpPath)
	if err != nil { http.Error(w, "queue error", http.StatusInternalServerError); return }

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
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
	for _, it := range items {
		row := respItem{
			VideoID:    strconv.Itoa(it.VideoID),
			Title:      it.Title,
			Status:     string(it.Status),
			UploadedAt: it.UploadedAt.Format(time.RFC3339),
		}
		if !it.ProcessedAt.IsZero() {
			row.ProcessedAt = it.ProcessedAt.Format(time.RFC3339)
		}
		if it.ProcessedURL != "" {
			row.ProcessedURL = it.ProcessedURL
		}
		out = append(out, row)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (h *VideosHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}
	v, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repos.ErrNotFound) {
			http.Error(w, "video no encontrado", http.StatusNotFound); return
		}
		http.Error(w, "video no encontrado", http.StatusNotFound); return
	}
	_ = json.NewEncoder(w).Encode(v)
}

func (h *VideosHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	if idStr == "" { idStr = r.URL.Query().Get("id") }
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 { http.Error(w, "id inválido", http.StatusBadRequest); return }

	if err := h.svc.Delete(r.Context(), id); err != nil {
		http.Error(w, "error al eliminar video", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "El video ha sido eliminado exitosamente.", "video_id": idStr})
}
