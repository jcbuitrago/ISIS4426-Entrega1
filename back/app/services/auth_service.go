package services

import (
	"context"
	"errors"
	"os"
	"time"

	"ISIS4426-Entrega1/app/models"

	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
)

type UserRepo interface {
	Create(ctx context.Context, u models.User) (models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id int) (*models.User, error)
	UpdateProfile(ctx context.Context, id int, firstName, lastName, city, country string) error
	UpdateAvatar(ctx context.Context, id int, url string) error
}

type AuthService struct{ users UserRepo }

func NewAuthService(r UserRepo) *AuthService { return &AuthService{users: r} }

// Expose underlying repo for profile handlers
func (a *AuthService) Users() UserRepo { return a.users }

var (
	ErrEmailExists      = errors.New("email ya registrado")
	ErrPasswordsNoMatch = errors.New("las contraseñas no coinciden")
	ErrInvalidCreds     = errors.New("credenciales inválidas")
)

func (a *AuthService) Signup(ctx context.Context, first, last, email, city, country, password1, password2 string) (models.User, error) {
	if password1 != password2 {
		return models.User{}, ErrPasswordsNoMatch
	}
	// si ya existe, devuelve ErrEmailExists (ignoramos el detalle del repo)
	if _, err := a.users.GetByEmail(ctx, email); err == nil {
		return models.User{}, ErrEmailExists
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password1), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}
	u := models.User{
		FirstName:    first,
		LastName:     last,
		Email:        email,
		City:         city,
		Country:      country,
		PasswordHash: string(hash),
	}
	return a.users.Create(ctx, u)
}

func (a *AuthService) Login(ctx context.Context, email, password string, remember bool) (string, time.Time, error) {
	u, err := a.users.GetByEmail(ctx, email)
	if err != nil {
		return "", time.Time{}, ErrInvalidCreds
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return "", time.Time{}, ErrInvalidCreds
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "devsecret123"
	}
	// session TTL: 1h default, 7d if remember
	ttl := time.Hour
	if remember {
		ttl = 7 * 24 * time.Hour
	}
	exp := time.Now().Add(ttl)
	claims := jwt.MapClaims{
		"user_id": u.ID,
		"email":   u.Email,
		"exp":     exp.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, err
	}
	return ss, exp, nil
}
