package usermodel

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUsernameRequired        = errors.New("username is required")
	ErrEmailRequired           = errors.New("email is required")
	ErrPasswordRequired        = errors.New("password is required")
	ErrActiveRequired          = errors.New("active is required")
	ErrCreatedAtRequired       = errors.New("created_at is required")
	ErrUpdatedAtRequired       = errors.New("updated_at is required")
	ErrDeletedAtRequired       = errors.New("deleted_at is required")
	ErrIDRequired              = errors.New("id is required")
	ErrRoleRequired            = errors.New("role is required")
	ErrRoleIDRequired          = errors.New("role_id is required")
	ErrRoleNameRequired        = errors.New("role_name is required")
	ErrRoleDescriptionRequired = errors.New("role_description is required")
	ErrRoleCreatedAtRequired   = errors.New("role_created_at is required")
	ErrRoleUpdatedAtRequired   = errors.New("role_updated_at is required")
)

type User struct {
	ID           uuid.UUID  `json:"id" validate:"required"`
	Username     string     `json:"username" validate:"required"`
	Email        string     `json:"email" validate:"required,email"`
	PasswordHash string     `json:"-" validate:"required"`
	Active       bool       `json:"active" validate:"required"`
	CreatedAt    time.Time  `json:"created_at" validate:"required,datetime"`
	UpdatedAt    time.Time  `json:"updated_at" validate:"required,datetime"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" validate:"omitempty,datetime"`
}

type UserRole struct {
	ID   uuid.UUID `json:"id" validate:"required"`
	Role string    `json:"role" validate:"required"`
}
