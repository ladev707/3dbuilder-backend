package main

import (
	"fmt"
	"os"
	"github.com/joho/godotenv"
	"github.com/ladev707/3dbuilder-backend/pkg/logger"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
	fmt.Println("app name is: ", os.Getenv("APP_NAME"))
	logger.Info("app started successfully")
}
