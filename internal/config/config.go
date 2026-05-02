package config

import "os"

type Config struct {
	Port         string
	DatabaseDSN  string
	JWTSecret    string
	MigrationDir string
}

func Load() Config {
	return Config{
		Port:         getenv("PORT", "8080"),
		DatabaseDSN:  getenv("DATABASE_DSN", "postgres://postgres:poopies@db:5432/exchange?sslmode=disable"),
		JWTSecret:    getenv("JWT_SECRET", "secret"),
		MigrationDir: getenv("MIGRATIONS_DIR", "migrations"),
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
