package userrepository

import (
	"context"
	"fmt"
)

func GetUser(ctx context.Context, userID string) {
	fmt.Println("GetUser")
}
