package userhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	usermodel "github.com/ladev707/3dbuilder-backend/internal/src/user/models"
)

// Handler is the HTTP adapter for user-management use cases.
// Authentication endpoints live in the auth module.
type Handler struct{}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) GetUser(c *gin.Context) {}

func (h *Handler) UpdateUser(c *gin.Context) {}

func (h *Handler) DeleteUser(c *gin.Context) {}

func (h *Handler) ListUsers(c *gin.Context) {}

func (h *Handler) CreateUser(c *gin.Context) {
	var user usermodel.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validator.New().Struct(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}
