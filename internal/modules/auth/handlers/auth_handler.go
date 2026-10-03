package authhandler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	authmiddleware "github.com/ladev707/3dbuilder-backend/internal/src/auth/middleware"
	authmodel "github.com/ladev707/3dbuilder-backend/internal/src/auth/models"
	authservice "github.com/ladev707/3dbuilder-backend/internal/src/auth/service"
)

type Handler struct {
	service *authservice.Service
}

func New(service *authservice.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Login(gate authmodel.Gate) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request authmodel.LoginRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "a valid email and password are required"})
			return
		}
		response, err := h.service.Login(c.Request.Context(), gate, request)
		if err != nil {
			switch {
			case errors.Is(err, authservice.ErrInvalidCredentials):
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			case errors.Is(err, authservice.ErrInactivePrincipal):
				c.JSON(http.StatusForbidden, gin.H{"error": "account is inactive"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "authentication failed"})
			}
			return
		}
		c.JSON(http.StatusOK, response)
	}
}

func (h *Handler) Refresh(gate authmodel.Gate) gin.HandlerFunc {
	return func(c *gin.Context) {
		request, ok := bindRefreshRequest(c)
		if !ok {
			return
		}
		response, err := h.service.Refresh(c.Request.Context(), gate, request.RefreshToken)
		if err != nil {
			if errors.Is(err, authservice.ErrInvalidToken) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token refresh failed"})
			return
		}
		c.JSON(http.StatusOK, response)
	}
}

func (h *Handler) Logout(gate authmodel.Gate) gin.HandlerFunc {
	return func(c *gin.Context) {
		request, ok := bindRefreshRequest(c)
		if !ok {
			return
		}
		if err := h.service.Logout(c.Request.Context(), gate, request.RefreshToken); err != nil {
			if errors.Is(err, authservice.ErrInvalidToken) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "logout failed"})
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func (h *Handler) Me(c *gin.Context) {
	claims, ok := authmiddleware.ClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	principal, err := h.service.CurrentPrincipal(c.Request.Context(), claims)
	if err != nil {
		if errors.Is(err, authservice.ErrInvalidToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "account is unavailable"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load account"})
		return
	}
	c.JSON(http.StatusOK, principal)
}

func bindRefreshRequest(c *gin.Context) (authmodel.RefreshRequest, bool) {
	var request authmodel.RefreshRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token is required"})
		return authmodel.RefreshRequest{}, false
	}
	return request, true
}
