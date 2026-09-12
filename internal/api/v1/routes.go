package v1

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/users", getUsers)
	rg.POST("/users", createUser)
}
