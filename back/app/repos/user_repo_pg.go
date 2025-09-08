package repos

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"ISIS4426-Entrega1/app/models"
)

type UserRepoPG struct{ DB *sql.DB }

func NewUserRepoPG(db *sql.DB) *UserRepoPG { return &UserRepoPG{DB: db} }

var ErrUserExists = errors.New("user already exists")
var ErrUserNotFound = errors.New("user not found")

func (r *UserRepoPG) Create(ctx context.Context, u models.User) (models.User, error) {
	const q = `
	INSERT INTO users (first_name, last_name, city, country, email, password_hash, created_at)
	VALUES ($1,$2,$3,$4,$5,$6,$7)
	RETURNING id, created_at`
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err := r.DB.QueryRowContext(ctx, q,
		u.FirstName, u.LastName, u.City, u.Country, u.Email, u.PasswordHash, time.Now(),
	).Scan(&u.ID, &u.CreatedAt)
	return u, err
}

func (r *UserRepoPG) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	const q = `SELECT id, first_name, last_name, city, country, email, password_hash, created_at FROM users WHERE email=$1`
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var u models.User
	err := r.DB.QueryRowContext(ctx, q, email).Scan(&u.ID, &u.FirstName, &u.LastName, &u.City, &u.Country, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
