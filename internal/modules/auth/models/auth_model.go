package authmodel

import (
	"time"

	"github.com/google/uuid"
)

type Gate string

const (
	GateUser     Gate = "user"
	GateCustomer Gate = "customer"
)

func (g Gate) Valid() bool {
	return g == GateUser || g == GateCustomer
}

type Principal struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Active       bool      `json:"-"`
	Gate         Gate      `json:"gate"`
	Roles        []string  `json:"roles"`
	Permissions  []string  `json:"permissions"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type LoginResponse struct {
	Tokens    TokenPair `json:"tokens"`
	Principal Principal `json:"principal"`
}

type Session struct {
	ID               uuid.UUID
	PrincipalID      uuid.UUID
	Gate             Gate
	RefreshTokenHash string
	ExpiresAt        time.Time
	RevokedAt        *time.Time
}

type Role struct {
	ID          uuid.UUID `json:"id"`
	Gate        Gate      `json:"gate"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

type Permission struct {
	ID          uuid.UUID `json:"id"`
	Gate        Gate      `json:"gate"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=80"`
	Description string `json:"description" binding:"max=500"`
}

type AssignmentRequest struct {
	RoleID       uuid.UUID `json:"role_id" binding:"required"`
	PermissionID uuid.UUID `json:"permission_id"`
	PrincipalID  uuid.UUID `json:"principal_id"`
}
