package user_dtos

import (
	"time"

	"github.com/google/uuid"
	usermodel "github.com/ladev707/3dbuilder-backend/internal/modules/user/models"
)

type CreateUserDTO struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func (d CreateUserDTO) ToModel(passwordHash string) usermodel.User {
	now := time.Now().UTC()
	return usermodel.User{
		ID:           uuid.New(),
		Username:     d.Username,
		Email:        d.Email,
		PasswordHash: passwordHash,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

type UserResponseDTO struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewUserResponseDTO(u usermodel.User) UserResponseDTO {
	return UserResponseDTO{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
