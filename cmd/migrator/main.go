package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"

	"github.com/ladev707/3dbuilder-backend/migrations"
)

func main() {
	_ = godotenv.Load()
	databaseURL := os.Getenv("DB_URL")
	if databaseURL == "" {
		log.Fatal("DB_URL is required")
	}
	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	goose.SetBaseFS(migrations.Files)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("set migration dialect: %v", err)
	}
	if err := goose.Run(command, db, "."); err != nil {
		log.Fatalf("goose %s: %v", command, err)
	}
}
