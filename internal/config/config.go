package config

import (
	"os"
	"time"
)

// Config holds all application configuration
type Config struct {
	Port        string
	DatabaseURL string
	RedisURL    string
	JWTSecret   string
	JWTExpiry   time.Duration
	Environment string
}

// Load reads configuration from environment variables
func Load() *Config {
	cfg := &Config{
		Port:        getEnv("PORT", ":8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:password@localhost/openbpl?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-key"),
		JWTExpiry:   parseDuration(getEnv("JWT_EXPIRY", "15m")),
		Environment: getEnv("ENVIRONMENT", "development"),
	}

	// Validate critical production settings
	if cfg.IsProduction() && cfg.JWTSecret == "dev-secret-key" {
		panic("JWT_SECRET must be set in production")
	}

	return cfg
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// parseDuration parses duration string with fallback
func parseDuration(s string) time.Duration {
	duration, err := time.ParseDuration(s)
	if err != nil {
		return 15 * time.Minute // default fallback
	}
	return duration
}

// IsDevelopment checks if we're in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction checks if we're in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}
