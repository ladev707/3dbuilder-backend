package usermiddleware

import (
	"github.com/gin-gonic/gin"
	userservice "github.com/ladev707/3dbuilder-backend/internal/src/user/service"
)

func UserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		userservice.GetUser(c.Request.Context(), userID)
		c.Next()
	}
}
