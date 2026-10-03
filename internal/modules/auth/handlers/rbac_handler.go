package authhandler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authmodel "github.com/ladev707/3dbuilder-backend/internal/src/auth/models"
	authrepository "github.com/ladev707/3dbuilder-backend/internal/src/auth/repository"
)

func (h *Handler) ListRoles(c *gin.Context) {
	gate, ok := targetGate(c)
	if !ok {
		return
	}
	roles, err := h.service.ListRoles(c.Request.Context(), gate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list roles"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

func (h *Handler) ListPermissions(c *gin.Context) {
	gate, ok := targetGate(c)
	if !ok {
		return
	}
	permissions, err := h.service.ListPermissions(c.Request.Context(), gate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list permissions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"permissions": permissions})
}

func (h *Handler) CreateRole(c *gin.Context) {
	gate, ok := targetGate(c)
	if !ok {
		return
	}
	var request authmodel.CreateRoleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valid role name is required"})
		return
	}
	role, err := h.service.CreateRole(c.Request.Context(), gate, request)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "role could not be created"})
		return
	}
	c.JSON(http.StatusCreated, role)
}

func (h *Handler) AssignPermission(c *gin.Context) {
	gate, roleID, permissionID, ok := assignmentIDs(c, "permission_id")
	if !ok {
		return
	}
	if err := h.service.AssignPermission(c.Request.Context(), gate, roleID, permissionID); err != nil {
		assignmentError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) RemovePermission(c *gin.Context) {
	gate, roleID, permissionID, ok := assignmentIDs(c, "permission_id")
	if !ok {
		return
	}
	if err := h.service.RemovePermission(c.Request.Context(), gate, roleID, permissionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "permission could not be removed"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) AssignRole(c *gin.Context) {
	gate, principalID, roleID, ok := assignmentIDs(c, "role_id")
	if !ok {
		return
	}
	if err := h.service.AssignRole(c.Request.Context(), gate, principalID, roleID); err != nil {
		assignmentError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) RemoveRole(c *gin.Context) {
	gate, principalID, roleID, ok := assignmentIDs(c, "role_id")
	if !ok {
		return
	}
	if err := h.service.RemoveRole(c.Request.Context(), gate, principalID, roleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "role could not be removed"})
		return
	}
	c.Status(http.StatusNoContent)
}

func targetGate(c *gin.Context) (authmodel.Gate, bool) {
	gate := authmodel.Gate(c.Param("gate"))
	if !gate.Valid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "gate must be user or customer"})
		return "", false
	}
	return gate, true
}

func assignmentIDs(c *gin.Context, secondParameter string) (authmodel.Gate, uuid.UUID, uuid.UUID, bool) {
	gate, ok := targetGate(c)
	if !ok {
		return "", uuid.Nil, uuid.Nil, false
	}
	firstParameter := "role_id"
	if secondParameter == "role_id" {
		firstParameter = "principal_id"
	}
	firstID, firstErr := uuid.Parse(c.Param(firstParameter))
	secondID, secondErr := uuid.Parse(c.Param(secondParameter))
	if firstErr != nil || secondErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resource IDs must be valid UUIDs"})
		return "", uuid.Nil, uuid.Nil, false
	}
	return gate, firstID, secondID, true
}

func assignmentError(c *gin.Context, err error) {
	if errors.Is(err, authrepository.ErrInvalidAssignment) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "assignment failed"})
}
