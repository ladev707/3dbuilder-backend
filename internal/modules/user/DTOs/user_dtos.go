package user_dtos

import usermodel "github.com/ladev707/3dbuilder-backend/internal/modules/user/models"

//get from model

type CreateUserDTO struct {
	User usermodel.User `json:"user" validate:"required"`
}

type UserResponseDTO struct {
	User usermodel.User `json:"user" validate:"required"`
}
