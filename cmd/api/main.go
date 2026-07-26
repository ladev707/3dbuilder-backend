package main

import (
	"os"
	"github.com/joho/godotenv"
	"github.com/ladev707/3dbuilder-backend/pkg/logger"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
	logger.InitLogger()
	logger.Info("app started successfully", "app_name", os.Getenv("APP_NAME"))
	logger.Debug("this is debug message", "key", "value")
	logger.Error("this is error message", "error", err)
}
