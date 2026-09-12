package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func getUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"users": []string{}})
}

func createUser(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"status": "created"})
}
