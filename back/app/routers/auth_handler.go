package routers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"ISIS4426-Entrega1/app/services"
)

type AuthHandler struct{ svc *services.AuthService }

func NewAuthHandler(svc *services.AuthService) *AuthHandler { return &AuthHandler{svc: svc} }

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
		http.Error(w, "json inválido", http.StatusBadRequest); return
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
	}
	var body req
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "json inválido", http.StatusBadRequest); return
	}
	tok, exp, err := h.svc.Login(r.Context(), body.Email, body.Password)
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
