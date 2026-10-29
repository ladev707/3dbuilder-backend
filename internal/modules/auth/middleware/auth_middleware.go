package authmiddleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ladev707/3dbuilder-backend/internal/config"
	authmodel "github.com/ladev707/3dbuilder-backend/internal/modules/auth/models"
	authservice "github.com/ladev707/3dbuilder-backend/internal/modules/auth/service"
)

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
		// Store the claims in two places:
		//   - the Gin context, read by Gin handlers via ClaimsFromContext;
		//   - the standard request context, read by code that only receives a
		//     context.Context (GraphQL resolvers and directives) via
		//     ClaimsFromRequestContext.
		// Contexts are immutable, so WithContext creates a copy of the request
		// carrying the new context.
		c.Set(config.ClaimsContextKey, claims)
		c.Request = c.Request.WithContext(ContextWithClaims(c.Request.Context(), claims))
		c.Next()
	}
}

// claimsContextKey is the key for storing claims in a context.Context.
// It is an unexported struct type rather than a string, so no other package
// can create the same key and overwrite or read the value by accident.
type claimsContextKey struct{}

// ContextWithClaims returns a copy of ctx that carries claims. It exposes
// the claims to handlers that only receive the request context, such as
// GraphQL resolvers.
func ContextWithClaims(ctx context.Context, claims *authservice.Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey{}, claims)
}

// ClaimsFromRequestContext returns the claims stored by Authenticate.
// ok is false when the request was not authenticated.
func ClaimsFromRequestContext(ctx context.Context) (*authservice.Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey{}).(*authservice.Claims)
	return claims, ok
}

func RequireGate(gate authmodel.Gate) gin.HandlerFunc {
	fmt.Println("RequireGate")
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
	value, exists := c.Get(config.ClaimsContextKey)
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
