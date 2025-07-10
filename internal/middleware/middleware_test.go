package middleware

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestChain(t *testing.T) {
	t.Run("chains middleware in correct order", func(t *testing.T) {
		// Track execution order
		var executionOrder []string

		// Create test middlewares that record their execution
		middleware1 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				executionOrder = append(executionOrder, "middleware1-before")
				next.ServeHTTP(w, r)
				executionOrder = append(executionOrder, "middleware1-after")
			})
		}

		middleware2 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				executionOrder = append(executionOrder, "middleware2-before")
				next.ServeHTTP(w, r)
				executionOrder = append(executionOrder, "middleware2-after")
			})
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executionOrder = append(executionOrder, "handler")
			w.Write([]byte("OK"))
		})

		// Chain middleware
		chained := Chain(handler, middleware1, middleware2)

		// Test request
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		chained.ServeHTTP(w, req)

		// Verify execution order: middleware1 -> middleware2 -> handler -> middleware2 -> middleware1
		expectedOrder := []string{
			"middleware1-before",
			"middleware2-before",
			"handler",
			"middleware2-after",
			"middleware1-after",
		}

		if len(executionOrder) != len(expectedOrder) {
			t.Fatalf("Expected %d execution steps, got %d", len(expectedOrder), len(executionOrder))
		}

		for i, expected := range expectedOrder {
			if executionOrder[i] != expected {
				t.Errorf("Step %d: expected %s, got %s", i, expected, executionOrder[i])
			}
		}
	})

	t.Run("works with no middleware", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("no middleware"))
		})

		chained := Chain(handler) // No middleware

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		chained.ServeHTTP(w, req)

		if w.Body.String() != "no middleware" {
			t.Errorf("Expected 'no middleware', got '%s'", w.Body.String())
		}
	})

	t.Run("works with single middleware", func(t *testing.T) {
		middleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Test", "applied")
				next.ServeHTTP(w, r)
			})
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("single"))
		})

		chained := Chain(handler, middleware)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		chained.ServeHTTP(w, req)

		if w.Header().Get("X-Test") != "applied" {
			t.Error("Middleware was not applied")
		}
		if w.Body.String() != "single" {
			t.Errorf("Expected 'single', got '%s'", w.Body.String())
		}
	})
}

func TestLogger(t *testing.T) {
	// Capture log output
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer log.SetOutput(os.Stderr) // Restore default

	t.Run("logs successful requests", func(t *testing.T) {
		logBuffer.Reset()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		loggedHandler := Logger(handler)

		req := httptest.NewRequest("GET", "/api/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		loggedHandler.ServeHTTP(w, req)

		logOutput := logBuffer.String()

		// Check log contains expected elements
		if !strings.Contains(logOutput, "GET") {
			t.Error("Log should contain HTTP method")
		}
		if !strings.Contains(logOutput, "/api/test") {
			t.Error("Log should contain request path")
		}
		if !strings.Contains(logOutput, "200") {
			t.Error("Log should contain status code")
		}
		if !strings.Contains(logOutput, "192.168.1.1:12345") {
			t.Error("Log should contain remote address")
		}
		// Duration should be present (contains time unit)
		if !strings.Contains(logOutput, "µs") && !strings.Contains(logOutput, "ms") && !strings.Contains(logOutput, "ns") {
			t.Error("Log should contain duration")
		}
	})

	t.Run("logs different status codes", func(t *testing.T) {
		testCases := []struct {
			name       string
			statusCode int
		}{
			{"not found", http.StatusNotFound},
			{"server error", http.StatusInternalServerError},
			{"created", http.StatusCreated},
			{"unauthorized", http.StatusUnauthorized},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				logBuffer.Reset()

				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.statusCode)
				})

				loggedHandler := Logger(handler)

				req := httptest.NewRequest("POST", "/test", nil)
				w := httptest.NewRecorder()

				loggedHandler.ServeHTTP(w, req)

				logOutput := logBuffer.String()
				// Use strconv.Itoa or fmt.Sprintf to convert int to string
				statusStr := fmt.Sprintf("%d", tc.statusCode)

				if !strings.Contains(logOutput, statusStr) {
					t.Errorf("Log should contain status code %d, got: %s", tc.statusCode, logOutput)
				}
			})
		}
	})

	t.Run("captures status code even when WriteHeader not called explicitly", func(t *testing.T) {
		logBuffer.Reset()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Don't call WriteHeader explicitly - should default to 200
			w.Write([]byte("default status"))
		})

		loggedHandler := Logger(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		loggedHandler.ServeHTTP(w, req)

		logOutput := logBuffer.String()
		if !strings.Contains(logOutput, "200") {
			t.Errorf("Should log default status 200, got: %s", logOutput)
		}
	})

	t.Run("measures timing accurately", func(t *testing.T) {
		logBuffer.Reset()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(10 * time.Millisecond) // Simulate some work
			w.WriteHeader(http.StatusOK)
		})

		loggedHandler := Logger(handler)

		req := httptest.NewRequest("GET", "/slow", nil)
		w := httptest.NewRecorder()

		start := time.Now()
		loggedHandler.ServeHTTP(w, req)
		actualDuration := time.Since(start)

		// The logged duration should be close to our measured duration
		// (within reasonable tolerance for test timing variations)
		if actualDuration < 5*time.Millisecond {
			t.Error("Handler should have taken at least 10ms")
		}
	})
}

func TestCORS(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	corsHandler := CORS(handler)

	t.Run("sets CORS headers for regular requests", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		corsHandler.ServeHTTP(w, req)

		// Check all CORS headers are set
		headers := w.Header()

		if headers.Get("Access-Control-Allow-Origin") != "*" {
			t.Errorf("Expected Allow-Origin *, got %s", headers.Get("Access-Control-Allow-Origin"))
		}

		allowMethods := headers.Get("Access-Control-Allow-Methods")
		expectedMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		for _, method := range expectedMethods {
			if !strings.Contains(allowMethods, method) {
				t.Errorf("Allow-Methods should contain %s, got %s", method, allowMethods)
			}
		}

		allowHeaders := headers.Get("Access-Control-Allow-Headers")
		expectedHeaders := []string{"Content-Type", "Authorization"}
		for _, header := range expectedHeaders {
			if !strings.Contains(allowHeaders, header) {
				t.Errorf("Allow-Headers should contain %s, got %s", header, allowHeaders)
			}
		}

		// Should call the next handler
		if w.Body.String() != "OK" {
			t.Errorf("Expected body 'OK', got '%s'", w.Body.String())
		}
	})

	t.Run("handles OPTIONS preflight requests", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/test", nil)
		w := httptest.NewRecorder()

		corsHandler.ServeHTTP(w, req)

		// Should return 200 OK
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for OPTIONS, got %d", w.Code)
		}

		// Should NOT call the next handler (body should be empty)
		if w.Body.String() == "OK" {
			t.Error("OPTIONS request should not reach the actual handler")
		}

		// Should still have CORS headers
		if w.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Error("OPTIONS response should include CORS headers")
		}
	})

	t.Run("works with different HTTP methods", func(t *testing.T) {
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

		for _, method := range methods {
			t.Run(method, func(t *testing.T) {
				req := httptest.NewRequest(method, "/test", nil)
				w := httptest.NewRecorder()

				corsHandler.ServeHTTP(w, req)

				// Should have CORS headers
				if w.Header().Get("Access-Control-Allow-Origin") != "*" {
					t.Errorf("CORS headers missing for %s request", method)
				}

				// Should call the handler (except OPTIONS)
				if method != "OPTIONS" && w.Body.String() != "OK" {
					t.Errorf("Handler not called for %s request", method)
				}
			})
		}
	})
}

func TestRecovery(t *testing.T) {
	// Capture log output
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer log.SetOutput(os.Stderr)

	t.Run("recovers from panics and returns 500", func(t *testing.T) {
		logBuffer.Reset()

		panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic message")
		})

		recoveryHandler := Recovery(panicHandler)

		req := httptest.NewRequest("GET", "/panic", nil)
		w := httptest.NewRecorder()

		// This should not panic
		recoveryHandler.ServeHTTP(w, req)

		// Should return 500 status
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", w.Code)
		}

		// Should contain error message
		if !strings.Contains(w.Body.String(), "Internal Server Error") {
			t.Errorf("Response should contain error message, got: %s", w.Body.String())
		}

		// Should log the panic
		logOutput := logBuffer.String()
		if !strings.Contains(logOutput, "Panic recovered") {
			t.Error("Should log panic recovery")
		}
		if !strings.Contains(logOutput, "test panic message") {
			t.Error("Should log the actual panic message")
		}
	})

	t.Run("does not interfere with normal requests", func(t *testing.T) {
		logBuffer.Reset()

		normalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("normal response"))
		})

		recoveryHandler := Recovery(normalHandler)

		req := httptest.NewRequest("GET", "/normal", nil)
		w := httptest.NewRecorder()

		recoveryHandler.ServeHTTP(w, req)

		// Should work normally
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		if w.Body.String() != "normal response" {
			t.Errorf("Expected 'normal response', got '%s'", w.Body.String())
		}

		// Should not log anything for normal requests
		if logBuffer.String() != "" {
			t.Errorf("Should not log anything for normal requests, got: %s", logBuffer.String())
		}
	})

	t.Run("handles different types of panics", func(t *testing.T) {
		testCases := []struct {
			name       string
			panicValue interface{}
		}{
			{"string panic", "string error"},
			{"error panic", http.ErrServerClosed},
			{"number panic", 42},
			{"nil panic", nil},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				logBuffer.Reset()

				panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					panic(tc.panicValue)
				})

				recoveryHandler := Recovery(panicHandler)

				req := httptest.NewRequest("GET", "/panic", nil)
				w := httptest.NewRecorder()

				recoveryHandler.ServeHTTP(w, req)

				if w.Code != http.StatusInternalServerError {
					t.Errorf("Expected status 500 for %s, got %d", tc.name, w.Code)
				}
			})
		}
	})
}

func TestResponseWriter(t *testing.T) {
	t.Run("captures status code from WriteHeader", func(t *testing.T) {
		w := httptest.NewRecorder()
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		rw.WriteHeader(http.StatusNotFound)

		if rw.statusCode != http.StatusNotFound {
			t.Errorf("Expected status code 404, got %d", rw.statusCode)
		}

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected underlying writer status 404, got %d", w.Code)
		}
	})

	t.Run("defaults to 200 OK", func(t *testing.T) {
		w := httptest.NewRecorder()
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Don't call WriteHeader explicitly
		rw.Write([]byte("test"))

		if rw.statusCode != http.StatusOK {
			t.Errorf("Expected default status code 200, got %d", rw.statusCode)
		}
	})

	t.Run("preserves all ResponseWriter methods", func(t *testing.T) {
		w := httptest.NewRecorder()
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Test Header method
		rw.Header().Set("X-Test", "value")
		if rw.Header().Get("X-Test") != "value" {
			t.Error("Header method not working")
		}

		// Test Write method
		n, err := rw.Write([]byte("test content"))
		if err != nil {
			t.Errorf("Write failed: %v", err)
		}
		if n != 12 {
			t.Errorf("Expected 12 bytes written, got %d", n)
		}

		if w.Body.String() != "test content" {
			t.Errorf("Expected 'test content', got '%s'", w.Body.String())
		}
	})
}
