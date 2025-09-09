package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type ctxKey string

const userIDKey ctxKey = "user_id"

func UserIDFromContext(ctx context.Context) (int, bool) {
	v := ctx.Value(userIDKey)
	id, ok := v.(int)
	return id, ok
}

func AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		parts := strings.Fields(auth) // separa por espacios múltiples
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "token faltante", http.StatusUnauthorized)
			return
		}
		tokenStr := parts[1]
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "devsecret123"
		}
		tok, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
			return []byte(secret), nil
		}, jwt.WithValidMethods([]string{"HS256"}))
		if err != nil || !tok.Valid {
			http.Error(w, "token inválido", http.StatusUnauthorized)
			return
		}
		claims, ok := tok.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "token inválido", http.StatusUnauthorized)
			return
		}
		// expiry
		if expVal, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(expVal) {
				http.Error(w, "token expirado", http.StatusUnauthorized)
				return
			}
		}
		uidFloat, ok := claims["user_id"].(float64)
		if !ok {
			http.Error(w, "token inválido", http.StatusUnauthorized)
			return
		}
		uid := int(uidFloat)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userIDKey, uid)))
	})
}
