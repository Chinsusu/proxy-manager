package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL      string
	JWTSecret        string
	APIBind          string
	AdminEmail       string
	AdminPassword    string
	JWTExpiration    time.Duration
}

func Load() *Config {
	return &Config{
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://pgm:pgm_password@localhost:5432/pgm_db?sslmode=disable"),
		JWTSecret:     getEnv("API_JWT_SECRET", "change_me_secret"),
		APIBind:       getEnv("API_BIND", ":8082"),
		AdminEmail:    getEnv("API_ADMIN_EMAIL", "admin@example.com"),
		AdminPassword: getEnv("API_ADMIN_PASSWORD", "admin_password"),
		JWTExpiration: time.Hour * time.Duration(getEnvAsInt("JWT_EXPIRATION_HOURS", 1)),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}
