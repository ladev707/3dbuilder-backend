package customermodel

import (
	"github.com/google/uuid"
	"time"
)

type Customer struct {
	ID           uuid.UUID  `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Active       bool       `json:"active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type ClientRole struct {
	ID   uuid.UUID `json:"id"`
	Role string    `json:"role"`
}
