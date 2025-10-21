package routers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"ISIS4426-Entrega1/app/middleware"
	"ISIS4426-Entrega1/app/services"
	"ISIS4426-Entrega1/internal/s3client"
)

type AuthHandler struct {
	svc      *services.AuthService
	s3Client *s3client.S3Client
}

func NewAuthHandler(svc *services.AuthService, s3 *s3client.S3Client) *AuthHandler {
	return &AuthHandler{svc: svc, s3Client: s3}
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	type req struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Password1 string `json:"password1"`
		Password2 string `json:"password2"`
		City      string `json:"city"`
		Country   string `json:"country"`
	}
	var body req
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "json inválido", http.StatusBadRequest)
		return
	}
	u, err := h.svc.Signup(r.Context(), body.FirstName, body.LastName, body.Email, body.City, body.Country, body.Password1, body.Password2)
	if err != nil {
		switch err {
		case services.ErrPasswordsNoMatch:
			http.Error(w, "contraseñas no coinciden", http.StatusBadRequest)
		case services.ErrEmailExists:
			http.Error(w, "email ya registrado", http.StatusBadRequest)
		default:
			log.Printf("signup error: %v", err)
			http.Error(w, "error interno", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":         u.ID,
		"first_name": u.FirstName,
		"last_name":  u.LastName,
		"email":      u.Email,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	type req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Remember bool   `json:"remember"`
	}
	var body req
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "json inválido", http.StatusBadRequest)
		return
	}
	tok, exp, err := h.svc.Login(r.Context(), body.Email, body.Password, body.Remember)
	if err != nil {
		if err == services.ErrInvalidCreds {
			http.Error(w, "credenciales inválidas", http.StatusUnauthorized)
			return
		}
		log.Printf("login error: %v", err)
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access_token": tok,
		"token_type":   "Bearer",
		"expires_in":   int(time.Until(exp).Seconds()),
	})
}

// GET /api/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	u, err := h.svc.Users().GetByID(r.Context(), uid)
	if err != nil {
		http.Error(w, "no encontrado", http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(u)
}

// PUT /api/me
func (h *AuthHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	type req struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		City      string `json:"city"`
		Country   string `json:"country"`
	}
	var body req
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "json inválido", http.StatusBadRequest)
		return
	}
	if err := h.svc.Users().UpdateProfile(r.Context(), uid, body.FirstName, body.LastName, body.City, body.Country); err != nil {
		http.Error(w, "error al actualizar", http.StatusInternalServerError)
		return
	}
	u, _ := h.svc.Users().GetByID(r.Context(), uid)
	_ = json.NewEncoder(w).Encode(u)
}

// POST /api/me/avatar (multipart form: avatar)
func (h *AuthHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(8 << 20); err != nil {
		http.Error(w, "multipart parse error", http.StatusBadRequest)
		return
	}

	f, hdr, err := r.FormFile("avatar")
	if err != nil {
		http.Error(w, "archivo faltante", http.StatusBadRequest)
		return
	}
	defer f.Close()

	// Generate S3 key for avatar
	s3Key := fmt.Sprintf("avatars/%d/%s", uid, filepath.Base(hdr.Filename))

	// Upload to S3 processed bucket (avatars should be publicly accessible)
	if err := h.s3Client.UploadToProcessed(r.Context(), s3Key, f); err != nil {
		http.Error(w, "cannot upload avatar to S3", http.StatusInternalServerError)
		return
	}

	// Generate S3 URL for the avatar
	avatarURL := h.s3Client.GetProcessedFileURL(s3Key)

	if err := h.svc.Users().UpdateAvatar(r.Context(), uid, avatarURL); err != nil {
		// Clean up S3 upload if DB update fails
		_ = h.s3Client.DeleteFile(r.Context(), h.s3Client.GetProcessedBucket(), s3Key)
		http.Error(w, "error al actualizar", http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]string{"avatar_url": avatarURL})
}
