package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	t.Run("loads defaults in development", func(t *testing.T) {
		// Clean environment for this test
		cleanupEnv := setupCleanEnv()
		defer cleanupEnv()

		cfg := Load()

		if cfg.Port != ":8080" {
			t.Errorf("Expected default port :8080, got %s", cfg.Port)
		}

		if cfg.JWTSecret != "dev-secret-key" {
			t.Errorf("Expected default JWT secret, got %s", cfg.JWTSecret)
		}

		if cfg.JWTExpiry != 15*time.Minute {
			t.Errorf("Expected default JWT expiry 15m, got %v", cfg.JWTExpiry)
		}

		if !cfg.IsDevelopment() {
			t.Error("Expected development mode")
		}
	})

	t.Run("loads environment variables", func(t *testing.T) {
		cleanupEnv := setupCleanEnv()
		defer cleanupEnv()

		os.Setenv("PORT", ":9000")
		os.Setenv("JWT_SECRET", "test-secret")
		os.Setenv("JWT_EXPIRY", "30m")
		os.Setenv("ENVIRONMENT", "production")

		cfg := Load()

		if cfg.Port != ":9000" {
			t.Errorf("Expected port :9000, got %s", cfg.Port)
		}

		if cfg.JWTSecret != "test-secret" {
			t.Errorf("Expected JWT secret test-secret, got %s", cfg.JWTSecret)
		}

		if cfg.JWTExpiry != 30*time.Minute {
			t.Errorf("Expected JWT expiry 30m, got %v", cfg.JWTExpiry)
		}

		if !cfg.IsProduction() {
			t.Error("Expected production mode")
		}
	})

	t.Run("panics on production with default JWT secret", func(t *testing.T) {
		cleanupEnv := setupCleanEnv()
		defer cleanupEnv()

		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when using default JWT secret in production")
			}
		}()

		os.Setenv("ENVIRONMENT", "production")
		// JWT_SECRET is not set, so it uses default

		Load() // This should panic
	})
}

// Helper function to clean environment variables
func setupCleanEnv() func() {
	// Store original values
	originalVars := map[string]string{
		"PORT":        os.Getenv("PORT"),
		"JWT_SECRET":  os.Getenv("JWT_SECRET"),
		"JWT_EXPIRY":  os.Getenv("JWT_EXPIRY"),
		"ENVIRONMENT": os.Getenv("ENVIRONMENT"),
	}

	// Clear all config-related env vars
	for key := range originalVars {
		os.Unsetenv(key)
	}

	// Return cleanup function
	return func() {
		for key, value := range originalVars {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"15m", 15 * time.Minute},
		{"1h", 1 * time.Hour},
		{"30s", 30 * time.Second},
		{"invalid", 15 * time.Minute}, // fallback
		{"", 15 * time.Minute},        // fallback
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseDuration(tt.input)
			if result != tt.expected {
				t.Errorf("parseDuration(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	defer os.Unsetenv("TEST_VAR")

	t.Run("returns environment variable when set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "test_value")
		result := getEnv("TEST_VAR", "fallback")
		if result != "test_value" {
			t.Errorf("Expected test_value, got %s", result)
		}
	})

	t.Run("returns fallback when not set", func(t *testing.T) {
		result := getEnv("NON_EXISTENT_VAR", "fallback")
		if result != "fallback" {
			t.Errorf("Expected fallback, got %s", result)
		}
	})
}
