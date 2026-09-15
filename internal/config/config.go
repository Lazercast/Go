package config

import "os"

// Config holds all runtime configuration for the application, loaded
// from environment variables so it can be changed per environment
// without touching code (see .env.example).
type Config struct {
	ServerPort  string
	DatabaseURL string
}

// Load reads configuration from environment variables, applying
// sensible defaults for local development.
func Load() Config {
	return Config{
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/tasks?sslmode=disable"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
