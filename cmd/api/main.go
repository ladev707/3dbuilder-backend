package main

import (
	"os"
	"github.com/joho/godotenv"
	"github.com/ladev707/3dbuilder-backend/pkg/logger"
	app "github.com/ladev707/3dbuilder-backend/internal/api"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
	logger.InitLogger()
	app := app.NewApplication(app.Config{
		Port: os.Getenv("APP_PORT"),
		Db: app.DbConfig{
			Dsn: os.Getenv("DATABASE_URL"),
		},
	})
	router := app.SetupRoutes()
	router.Run(":" + os.Getenv("APP_PORT"))

	logger.Info("app started successfully", "app_name", os.Getenv("APP_NAME"))
	logger.Debug("this is debug message", "key", "value")
	logger.Error("this is error message", "error", err)
}
