package authmiddleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	authmodel "github.com/ladev707/3dbuilder-backend/internal/src/auth/models"
	authservice "github.com/ladev707/3dbuilder-backend/internal/src/auth/service"
)

const claimsContextKey = "auth.claims"

type TokenParser interface {
	ParseToken(string, string) (*authservice.Claims, error)
}

func Authenticate(parser TokenParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.Fields(header)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			abortUnauthorized(c)
			return
		}
		claims, err := parser.ParseToken(parts[1], "access")
		if err != nil {
			abortUnauthorized(c)
			return
		}
		c.Set(claimsContextKey, claims)
		c.Next()
	}
}

func RequireGate(gate authmodel.Gate) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := ClaimsFromContext(c)
		if !ok {
			abortUnauthorized(c)
			return
		}
		if claims.Gate != gate {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "this authentication gate cannot access the resource"})
			return
		}
		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return requireAny("required role is missing", roles, func(claims *authservice.Claims) []string {
		return claims.Roles
	})
}

func RequirePermission(permissions ...string) gin.HandlerFunc {
	return requireAny("required permission is missing", permissions, func(claims *authservice.Claims) []string {
		return claims.Permissions
	})
}

func ClaimsFromContext(c *gin.Context) (*authservice.Claims, bool) {
	value, exists := c.Get(claimsContextKey)
	if !exists {
		return nil, false
	}
	claims, ok := value.(*authservice.Claims)
	return claims, ok
}

func requireAny(message string, expected []string, values func(*authservice.Claims) []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := ClaimsFromContext(c)
		if !ok {
			abortUnauthorized(c)
			return
		}
		for _, actual := range values(claims) {
			for _, required := range expected {
				if actual == required {
					c.Next()
					return
				}
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": message})
	}
}

func abortUnauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
}
