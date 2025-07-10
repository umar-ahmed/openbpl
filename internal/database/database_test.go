package database

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func TestConnect(t *testing.T) {
	t.Run("successful connection with valid URL", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping database test in short mode")
		}

		testURL := getTestDatabaseURL()
		if testURL == "" {
			t.Skip("No test database URL provided")
		}

		db, err := Connect(testURL)
		if err != nil {
			t.Fatalf("Expected successful connection, got error: %v", err)
		}
		defer db.Close()

		if err := db.Ping(); err != nil {
			t.Errorf("Expected successful ping, got error: %v", err)
		}
	})

	t.Run("fails with invalid database URL", func(t *testing.T) {
		invalidURL := "postgres://invalid:invalid@nonexistent:5432/invalid"

		db, err := Connect(invalidURL)
		if err == nil {
			db.Close()
			t.Fatal("Expected connection to fail with invalid URL")
		}

		if err.Error() == "" {
			t.Error("Expected descriptive error message")
		}
	})

	t.Run("fails with malformed URL", func(t *testing.T) {
		malformedURL := "not-a-valid-url"

		_, err := Connect(malformedURL)
		if err == nil {
			t.Fatal("Expected connection to fail with malformed URL")
		}
	})

	t.Run("sets connection pool parameters", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping database test in short mode")
		}

		testURL := getTestDatabaseURL()
		if testURL == "" {
			t.Skip("No test database URL provided")
		}

		db, err := Connect(testURL)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer db.Close()

		stats := db.Stats()

		if stats.MaxOpenConnections <= 0 {
			t.Error("Expected MaxOpenConnections to be set")
		}

		if stats.MaxOpenConnections > 100 {
			t.Error("MaxOpenConnections seems too high for default config")
		}
	})
}

func TestDB_Close(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database test in short mode")
	}

	testURL := getTestDatabaseURL()
	if testURL == "" {
		t.Skip("No test database URL provided")
	}

	db, err := Connect(testURL)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Errorf("Expected Close() to succeed, got error: %v", err)
	}

	if err := db.Ping(); err == nil {
		t.Error("Expected Ping() to fail after Close()")
	}
}

func TestDB_HealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database test in short mode")
	}

	testURL := getTestDatabaseURL()
	if testURL == "" {
		t.Skip("No test database URL provided")
	}

	t.Run("healthy database returns no error", func(t *testing.T) {
		db, err := Connect(testURL)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer db.Close()

		if err := db.HealthCheck(); err != nil {
			t.Errorf("Expected healthy database, got error: %v", err)
		}
	})

	t.Run("health check respects context timeout", func(t *testing.T) {
		db, err := Connect(testURL)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer db.Close()

		start := time.Now()
		err = db.HealthCheck()
		duration := time.Since(start)

		if err != nil {
			t.Errorf("Health check failed: %v", err)
		}

		if duration > 1*time.Second {
			t.Errorf("Health check took too long: %v", duration)
		}
	})

	t.Run("health check fails on closed connection", func(t *testing.T) {
		db, err := Connect(testURL)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}

		db.Close()

		if err := db.HealthCheck(); err == nil {
			t.Error("Expected health check to fail on closed connection")
		}
	})
}

func getTestDatabaseURL() string {
	testURLs := []string{
		"postgres://openbpl_user:openbpl_password@localhost:5432/openbpl?sslmode=disable",
		"postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
		"postgres://user:password@localhost:5432/openbpl_test?sslmode=disable",
	}

	for _, url := range testURLs {
		if db, err := sql.Open("postgres", url); err == nil {
			if err := db.Ping(); err == nil {
				db.Close()
				return url
			}
			db.Close()
		}
	}

	return ""
}

func BenchmarkConnect(b *testing.B) {
	testURL := getTestDatabaseURL()
	if testURL == "" {
		b.Skip("No test database URL provided")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db, err := Connect(testURL)
		if err != nil {
			b.Fatalf("Connection failed: %v", err)
		}
		db.Close()
	}
}

func BenchmarkHealthCheck(b *testing.B) {
	testURL := getTestDatabaseURL()
	if testURL == "" {
		b.Skip("No test database URL provided")
	}

	db, err := Connect(testURL)
	if err != nil {
		b.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := db.HealthCheck(); err != nil {
			b.Fatalf("Health check failed: %v", err)
		}
	}
}
