package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	authaction "github.com/ladev707/3dbuilder-backend/internal/src/auth/actions"
	authmodel "github.com/ladev707/3dbuilder-backend/internal/src/auth/models"
)

func main() {
	_ = godotenv.Load()
	databaseURL := os.Getenv("DATABASE_URL")
	email := strings.TrimSpace(os.Getenv("BOOTSTRAP_EMAIL"))
	username := strings.TrimSpace(os.Getenv("BOOTSTRAP_USERNAME"))
	password := os.Getenv("BOOTSTRAP_PASSWORD")
	gate := authmodel.Gate(envOrDefault("BOOTSTRAP_GATE", string(authmodel.GateUser)))
	role := strings.TrimSpace(envOrDefault("BOOTSTRAP_ROLE", defaultRole(gate)))
	if databaseURL == "" || email == "" || username == "" || len(password) < 8 || !gate.Valid() || role == "" {
		log.Fatal("DATABASE_URL, valid BOOTSTRAP_GATE, BOOTSTRAP_EMAIL, BOOTSTRAP_USERNAME, BOOTSTRAP_ROLE, and BOOTSTRAP_PASSWORD (minimum 8 characters) are required")
	}

	passwordHash, err := authaction.HashPassword(password)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}
	ctx := context.Background()
	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer db.Close()

	table, joinTable, ownerColumn := gateTables(gate)
	tx, err := db.Begin(ctx)
	if err != nil {
		log.Fatalf("begin transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	var principalID uuid.UUID
	query := fmt.Sprintf(`
		INSERT INTO %s (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id`, table)
	if err := tx.QueryRow(ctx, query, username, email, passwordHash).Scan(&principalID); err != nil {
		log.Fatalf("create %s account: %v", gate, err)
	}
	query = fmt.Sprintf(`
		INSERT INTO %s (%s, role_id, gate)
		SELECT $1, id, $2::varchar FROM roles WHERE gate = $2::varchar AND name = $3`, joinTable, ownerColumn)
	tag, err := tx.Exec(ctx, query, principalID, gate, role)
	if err != nil {
		log.Fatalf("assign role: %v", err)
	}
	if tag.RowsAffected() != 1 {
		log.Fatalf("role %q does not exist for the %s gate", role, gate)
	}
	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("commit account: %v", err)
	}
	log.Printf("created %s account %s with role %s", gate, email, role)
}

func gateTables(gate authmodel.Gate) (string, string, string) {
	if gate == authmodel.GateCustomer {
		return "customers", "customer_roles", "customer_id"
	}
	return "users", "user_roles", "user_id"
}

func defaultRole(gate authmodel.Gate) string {
	if gate == authmodel.GateCustomer {
		return "customer"
	}
	return "admin"
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
