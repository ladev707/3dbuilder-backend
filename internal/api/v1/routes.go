package v1

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"

	authhandler "github.com/ladev707/3dbuilder-backend/internal/modules/auth/handlers"
	authmiddleware "github.com/ladev707/3dbuilder-backend/internal/modules/auth/middleware"
	authmodel "github.com/ladev707/3dbuilder-backend/internal/modules/auth/models"
	authservice "github.com/ladev707/3dbuilder-backend/internal/modules/auth/service"
)

type Dependencies struct {
	AuthHandler *authhandler.Handler
	AuthService *authservice.Service
	// GraphQLHandler serves the user and customer CRUD API (see internal/graph).
	GraphQLHandler http.Handler
	// EnablePlayground mounts the browser IDE at /graphql/playground.
	EnablePlayground bool
}

func RegisterRoutes(rg *gin.RouterGroup, dependencies Dependencies) {
	registerAuthGate(rg, authmodel.GateUser, dependencies)
	registerAuthGate(rg, authmodel.GateCustomer, dependencies)

	// GraphQL has a single endpoint for every operation. Authentication and
	// the "user" gate are checked here once; per-operation permissions are
	// checked by the @hasPermission directive in the schema. Customers' tokens
	// are rejected with 403 by RequireGate.
	graphql := rg.Group("/graphql")
	graphql.Use(
		authmiddleware.Authenticate(dependencies.AuthService),
		authmiddleware.RequireGate(authmodel.GateUser),
	)
	// gin.WrapH adapts a standard http.Handler (what gqlgen builds) into a
	// Gin handler.
	graphql.GET("", gin.WrapH(dependencies.GraphQLHandler))
	graphql.POST("", gin.WrapH(dependencies.GraphQLHandler))
	// The Playground page itself is public static HTML; the queries it sends
	// still need a token (set it under "Headers" in the Playground UI).
	if dependencies.EnablePlayground {
		rg.GET("/graphql/playground", gin.WrapH(playground.Handler("3D Builder GraphQL", rg.BasePath()+"/graphql")))
	}

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
