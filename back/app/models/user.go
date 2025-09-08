package models

import "time"

type User struct {
	ID           int       `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `json:"email"`
	City         string    `json:"city"`
	Country      string    `json:"country"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}
