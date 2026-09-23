package userservice

import (
	"context"

	userrepository "github.com/ladev707/3dbuilder-backend/internal/src/user/repository"
)

func GetUser(ctx context.Context, userID string) {
	userrepository.GetUser(ctx, userID)
}
