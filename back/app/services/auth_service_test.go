package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"ISIS4426-Entrega1/app/models"
)

type mockUserRepo struct {
	users map[string]models.User
}

func (m *mockUserRepo) Create(ctx context.Context, u models.User) (models.User, error) {
	m.users[u.Email] = u
	u.ID = len(m.users) // simular autoincrement
	return u, nil
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	u, ok := m.users[email]
	if !ok {
		return nil, errors.New("not found")
	}
	return &u, nil
}
func (m *mockUserRepo) GetByID(ctx context.Context, id int) (*models.User, error) { return nil, nil }
func (m *mockUserRepo) UpdateProfile(ctx context.Context, id int, firstName, lastName, city, country string) error {
	return nil
}
func (m *mockUserRepo) UpdateAvatar(ctx context.Context, id int, url string) error { return nil }

func TestAuthService_Signup_PasswordsNoMatch(t *testing.T) {
	svc := NewAuthService(&mockUserRepo{users: map[string]models.User{}})
	_, err := svc.Signup(context.Background(), "A", "B", "a@b.com", "City", "CO", "pass1", "pass2")
	if err != ErrPasswordsNoMatch {
		t.Fatalf("expected ErrPasswordsNoMatch, got %v", err)
	}
}

func TestAuthService_Signup_EmailExists(t *testing.T) {
	repo := &mockUserRepo{users: map[string]models.User{"a@b.com": {Email: "a@b.com"}}}
	svc := NewAuthService(repo)
	_, err := svc.Signup(context.Background(), "A", "B", "a@b.com", "City", "CO", "pass", "pass")
	if err != ErrEmailExists {
		t.Fatalf("expected ErrEmailExists, got %v", err)
	}
}

func TestAuthService_Signup_Success(t *testing.T) {
	repo := &mockUserRepo{users: map[string]models.User{}}
	svc := NewAuthService(repo)
	u, err := svc.Signup(context.Background(), "A", "B", "a@b.com", "City", "CO", "secret", "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.PasswordHash == "" {
		t.Fatal("expected PasswordHash to be set")
	}
}

func TestAuthService_Login_InvalidUser(t *testing.T) {
	svc := NewAuthService(&mockUserRepo{users: map[string]models.User{}})
	_, _, err := svc.Login(context.Background(), "none@b.com", "pass", false)
	if err != ErrInvalidCreds {
		t.Fatalf("expected ErrInvalidCreds, got %v", err)
	}
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	repo := &mockUserRepo{users: map[string]models.User{}}
	svc := NewAuthService(repo)
	// crear usuario con password válido
	_, _ = svc.Signup(context.Background(), "A", "B", "a@b.com", "City", "CO", "secret", "secret")
	_, _, err := svc.Login(context.Background(), "a@b.com", "wrong", false)
	if err != ErrInvalidCreds {
		t.Fatalf("expected ErrInvalidCreds, got %v", err)
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	repo := &mockUserRepo{users: map[string]models.User{}}
	svc := NewAuthService(repo)
	u, _ := svc.Signup(context.Background(), "A", "B", "a@b.com", "City", "CO", "secret", "secret")

	token, exp, err := svc.Login(context.Background(), u.Email, "secret", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty JWT token")
	}
	if time.Until(exp) > time.Hour+time.Minute || time.Until(exp) < time.Hour-time.Minute {
		t.Fatalf("expected exp around 1h, got %v", exp)
	}
}

func TestAuthService_Login_RememberSuccess(t *testing.T) {
	repo := &mockUserRepo{users: map[string]models.User{}}
	svc := NewAuthService(repo)
	u, _ := svc.Signup(context.Background(), "A", "B", "a@b.com", "City", "CO", "secret", "secret")

	_, exp, err := svc.Login(context.Background(), u.Email, "secret", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if time.Until(exp) < 6*24*time.Hour {
		t.Fatalf("expected exp ~7d, got %v", exp)
	}
}
