package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	v1 "github.com/ladev707/3dbuilder-backend/internal/api/v1"
	"github.com/ladev707/3dbuilder-backend/internal/config"
	"github.com/ladev707/3dbuilder-backend/internal/graph"
	"github.com/ladev707/3dbuilder-backend/internal/graph/resolver"
	authhandler "github.com/ladev707/3dbuilder-backend/internal/modules/auth/handlers"
	authrepository "github.com/ladev707/3dbuilder-backend/internal/modules/auth/repository"
	authservice "github.com/ladev707/3dbuilder-backend/internal/modules/auth/service"
	customerrepository "github.com/ladev707/3dbuilder-backend/internal/modules/customer/repository"
	customerservice "github.com/ladev707/3dbuilder-backend/internal/modules/customer/service"
	userrepository "github.com/ladev707/3dbuilder-backend/internal/modules/user/repository"
	userservice "github.com/ladev707/3dbuilder-backend/internal/modules/user/service"
)

type Application struct {
	Config config.Config
	db     *pgxpool.Pool
}

func NewApplication(ctx context.Context, cfg config.Config) (*Application, error) {
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}
	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	return &Application{Config: cfg, db: db}, nil
}

func (app *Application) Close() {
	app.db.Close()
}

func (app *Application) SetupRoutes() *gin.Engine {
	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"The api is running": "ok"})
	})

	authRepository := authrepository.New(app.db)
	authService := authservice.New(authRepository, app.Config.Auth)
	authHandler := authhandler.New(authService)

	// Dependency wiring for GraphQL: repository (SQL) → service (business
	// rules) → resolver (GraphQL adapter). Each layer receives the one below
	// it, so they can be swapped or faked in tests independently. Nothing
	// here touches the database yet, which is why routes_test.go can build
	// the router with a nil pool.
	graphQLHandler := graph.NewHandler(&resolver.Resolver{
		UserService:     userservice.New(userrepository.New(app.db)),
		CustomerService: customerservice.New(customerrepository.New(app.db)),
	}, graph.Options{Introspection: !app.Config.IsProduction()})

	v1Group := router.Group("/api/v1")
	v1.RegisterRoutes(v1Group, v1.Dependencies{
		AuthHandler:      authHandler,
		AuthService:      authService,
		GraphQLHandler:   graphQLHandler,
		EnablePlayground: !app.Config.IsProduction(),
	})

	return router
}
