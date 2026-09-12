package api

import (
	"net/http"
	"github.com/gin-gonic/gin"

	v1 "github.com/ladev707/3dbuilder-backend/internal/api/v1"
)

type Application struct {
	Config Config
}

type Config struct {
	Port string
	Db   DbConfig
}

type DbConfig struct {
	Dsn string
}

func NewApplication(cfg Config) *Application {
	return &Application{Config: cfg}
}

func (app *Application) SetupRoutes() *gin.Engine {
	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"status": "ok"})
	})

	v1Group := router.Group("/api/v1")
	v1.RegisterRoutes(v1Group)

	// v2Group := router.Group("/api/v2")
	// v2.RegisterRoutes(v2Group)

	return router
}
