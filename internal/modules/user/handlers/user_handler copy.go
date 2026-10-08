package userhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	usermodel "github.com/ladev707/3dbuilder-backend/internal/modules/user/models"
	userservice "github.com/ladev707/3dbuilder-backend/internal/modules/user/service"
)

// Handler is the HTTP adapter for user-management use cases.
// Authentication endpoints live in the auth module.
type Handler struct {
	UserService *userservice.Service
}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	user, err := h.UserService.UserRepository.GetUser(c, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

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
