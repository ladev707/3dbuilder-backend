package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type AuthConfig struct {
	Secret          string
	Issuer          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type Config struct {
	AppPort     string
	DatabaseURL string
	Auth        AuthConfig
}

func Load() (Config, error) {
	accessMinutes, err := positiveIntEnv("AUTH_ACCESS_TOKEN_MINUTES", 15)
	if err != nil {
		return Config{}, err
	}
	refreshHours, err := positiveIntEnv("AUTH_REFRESH_TOKEN_HOURS", 720)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		AppPort:     envOrDefault("APP_PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Auth: AuthConfig{
			Secret:          os.Getenv("AUTH_JWT_SECRET"),
			Issuer:          envOrDefault("AUTH_JWT_ISSUER", "3dbuilder-api"),
			AccessTokenTTL:  time.Duration(accessMinutes) * time.Minute,
			RefreshTokenTTL: time.Duration(refreshHours) * time.Hour,
		},
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if len(cfg.Auth.Secret) < 32 {
		return Config{}, fmt.Errorf("AUTH_JWT_SECRET must be at least 32 characters")
	}
	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func positiveIntEnv(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	number, err := strconv.Atoi(value)
	if err != nil || number <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return number, nil
}
