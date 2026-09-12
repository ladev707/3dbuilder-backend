package v1

import (
	"github.com/gin-gonic/gin"

	authhandler "github.com/ladev707/3dbuilder-backend/internal/src/auth/handlers"
	authmiddleware "github.com/ladev707/3dbuilder-backend/internal/src/auth/middleware"
	authmodel "github.com/ladev707/3dbuilder-backend/internal/src/auth/models"
	authservice "github.com/ladev707/3dbuilder-backend/internal/src/auth/service"
)

type Dependencies struct {
	AuthHandler *authhandler.Handler
	AuthService *authservice.Service
}

func RegisterRoutes(rg *gin.RouterGroup, dependencies Dependencies) {
	registerAuthGate(rg, authmodel.GateUser, dependencies)
	registerAuthGate(rg, authmodel.GateCustomer, dependencies)

	users := rg.Group("/users")
	users.Use(
		authmiddleware.Authenticate(dependencies.AuthService),
		authmiddleware.RequireGate(authmodel.GateUser),
	)
	users.GET("", authmiddleware.RequirePermission("users.read"), getUsers)
	users.POST("", authmiddleware.RequirePermission("users.create"), createUser)

	rbac := rg.Group("/rbac/:gate")
	rbac.Use(
		authmiddleware.Authenticate(dependencies.AuthService),
		authmiddleware.RequireGate(authmodel.GateUser),
		authmiddleware.RequirePermission("roles.manage"),
	)
	rbac.GET("/roles", dependencies.AuthHandler.ListRoles)
	rbac.POST("/roles", dependencies.AuthHandler.CreateRole)
	rbac.GET("/permissions", dependencies.AuthHandler.ListPermissions)
	rbac.PUT("/roles/:role_id/permissions/:permission_id", dependencies.AuthHandler.AssignPermission)
	rbac.DELETE("/roles/:role_id/permissions/:permission_id", dependencies.AuthHandler.RemovePermission)
	rbac.PUT("/principals/:principal_id/roles/:role_id", dependencies.AuthHandler.AssignRole)
	rbac.DELETE("/principals/:principal_id/roles/:role_id", dependencies.AuthHandler.RemoveRole)
}

func registerAuthGate(rg *gin.RouterGroup, gate authmodel.Gate, dependencies Dependencies) {
	group := rg.Group("/auth/" + string(gate))
	group.POST("/login", dependencies.AuthHandler.Login(gate))
	group.POST("/refresh", dependencies.AuthHandler.Refresh(gate))

	protected := group.Group("")
	protected.Use(
		authmiddleware.Authenticate(dependencies.AuthService),
		authmiddleware.RequireGate(gate),
	)
	protected.POST("/logout", dependencies.AuthHandler.Logout(gate))
	if gate == authmodel.GateCustomer {
		protected.GET("/me", authmiddleware.RequirePermission("profile.read"), dependencies.AuthHandler.Me)
		return
	}
	protected.GET("/me", dependencies.AuthHandler.Me)
}
