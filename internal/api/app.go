package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	v1 "github.com/ladev707/3dbuilder-backend/internal/api/v1"
	"github.com/ladev707/3dbuilder-backend/internal/config"
	authhandler "github.com/ladev707/3dbuilder-backend/internal/src/auth/handlers"
	authrepository "github.com/ladev707/3dbuilder-backend/internal/src/auth/repository"
	authservice "github.com/ladev707/3dbuilder-backend/internal/src/auth/service"
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
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	authRepository := authrepository.New(app.db)
	authService := authservice.New(authRepository, app.Config.Auth)
	authHandler := authhandler.New(authService)

	v1Group := router.Group("/api/v1")
	v1.RegisterRoutes(v1Group, v1.Dependencies{
		AuthHandler: authHandler,
		AuthService: authService,
	})

	return router
}
