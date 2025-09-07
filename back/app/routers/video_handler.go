package routers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"ISIS4426-Entrega1/app/async"
	"ISIS4426-Entrega1/app/repos"
	"ISIS4426-Entrega1/app/services"
	"strconv"

	"github.com/gorilla/mux"
)

type VideosHandler struct {
	svc      *services.VideoService
	enqueuer *async.Enqueuer
}

func NewVideosHandler(repo services.VideoRepo, enq *async.Enqueuer) *VideosHandler {
	return &VideosHandler{svc: services.NewVideoService(repo), enqueuer: enq}
}

func (h *VideosHandler) Create(w http.ResponseWriter, r *http.Request) {
	// 1) Limitar tamaño total (ej. 200MB)
	const maxUpload = 200 << 20 // 200 MiB
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)

	// 2) Parsear multipart (buffer en memoria 32MB; resto a tmp)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "multipart parse error", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 3) Ruta destino portable: <temp>/videos/<nombre>
	baseDir := filepath.Join(os.TempDir(), "videos")
	if err := os.MkdirAll(baseDir, 0o775); err != nil {
		http.Error(w, "cannot prepare storage dir", http.StatusInternalServerError)
		return
	}

	// Saneamos el nombre para evitar traversal y vacíos
	name := filepath.Base(header.Filename)
	if name == "" {
		http.Error(w, "invalid filename", http.StatusBadRequest)
		return
	}
	dstPath := filepath.Join(baseDir, name)

	// 4) Crear archivo de salida
	out, err := os.Create(dstPath)
	if err != nil {
		// TIP: revisa permisos/ruta si llegas aquí
		http.Error(w, "cannot save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	// 5) Copiar bytes
	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, "cannot write file", http.StatusInternalServerError)
		return
	}

	// 6) Persistir metadata (por ahora repo en memoria)
	v, err := h.svc.Create(1, title, dstPath) // userID=1 stub
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	payload := async.ProcessVideoPayload{
		VideoID: v.VideoID, SrcPath: v.OriginURL, UserID: v.UserID,
	}
	if _, err := h.enqueuer.EnqueueProcessVideo(r.Context(), payload); err != nil {
		// encolado falló: puedes loguear y dejar status "uploaded"
		// o marcar "failed_enqueued". Por simplicidad, devolvemos 202 con warning.
		// Pero para ahora: responde 201 normal y log interno.
		log.Println("failed to enqueue video processing:", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *VideosHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	v, err := h.svc.GetByID(id)

	if err != nil {
		status := http.StatusInternalServerError
		if err == repos.ErrNotFound {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func (h *VideosHandler) List(w http.ResponseWriter, r *http.Request) {
	videos, err := h.svc.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(videos)
}

func (h *VideosHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.svc.Delete(id); err != nil {
		status := http.StatusInternalServerError
		if err == repos.ErrNotFound {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
