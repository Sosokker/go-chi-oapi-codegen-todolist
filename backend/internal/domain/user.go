package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID `json:"id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	PasswordHash  string    `json:"-"`
	EmailVerified bool      `json:"emailVerified"`
	GoogleID      *string   `json:"-"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
