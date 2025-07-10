package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealth(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	Health(w, req)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	// Parse and check response
	var response Response
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response.Status)
	}

	if response.Message != "OpenBPL is running" {
		t.Errorf("Expected message 'OpenBPL is running', got '%s'", response.Message)
	}

	// Check that data field exists and has expected keys
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}

	if _, exists := data["timestamp"]; !exists {
		t.Error("Expected timestamp in data")
	}

	if _, exists := data["uptime"]; !exists {
		t.Error("Expected uptime in data")
	}
}

func TestStatus(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/status", nil)
	w := httptest.NewRecorder()

	Status(w, req)

	// Check basic response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response Response
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response.Status)
	}

	// Check data structure
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}

	expectedKeys := []string{"service", "version", "environment", "timestamp", "endpoints"}
	for _, key := range expectedKeys {
		if _, exists := data[key]; !exists {
			t.Errorf("Expected key '%s' in data", key)
		}
	}

	// Check service name
	if service, _ := data["service"].(string); service != "OpenBPL" {
		t.Errorf("Expected service 'OpenBPL', got '%s'", service)
	}
}

func TestHome(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	Home(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response Response
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response.Status)
	}

	if !strings.Contains(response.Message, "Welcome") {
		t.Errorf("Expected welcome message, got '%s'", response.Message)
	}

	// Check data has expected fields
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}

	if _, exists := data["description"]; !exists {
		t.Error("Expected description in data")
	}
}

func TestNotFound(t *testing.T) {
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()

	NotFound(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var response Response
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "error" {
		t.Errorf("Expected status 'error', got '%s'", response.Status)
	}

	if !strings.Contains(response.Error, "not found") {
		t.Errorf("Expected 'not found' in error, got '%s'", response.Error)
	}

	if !strings.Contains(response.Error, "/nonexistent") {
		t.Errorf("Expected path in error message, got '%s'", response.Error)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest("POST", "/health", nil)
	w := httptest.NewRecorder()

	MethodNotAllowed(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}

	var response Response
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "error" {
		t.Errorf("Expected status 'error', got '%s'", response.Status)
	}

	if !strings.Contains(response.Error, "POST") {
		t.Errorf("Expected method in error, got '%s'", response.Error)
	}

	if !strings.Contains(response.Error, "not allowed") {
		t.Errorf("Expected 'not allowed' in error, got '%s'", response.Error)
	}
}

func TestWriteJSONResponse(t *testing.T) {
	t.Run("writes valid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := Response{Status: "test", Message: "test message"}

		writeJSONResponse(w, http.StatusOK, data)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		if w.Header().Get("Content-Type") != "application/json" {
			t.Error("Expected application/json content type")
		}

		var response Response
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Status != "test" {
			t.Errorf("Expected status 'test', got '%s'", response.Status)
		}
	})

	t.Run("handles different status codes", func(t *testing.T) {
		testCases := []int{200, 201, 400, 404, 500}

		for _, statusCode := range testCases {
			w := httptest.NewRecorder()
			data := Response{Status: "test"}

			writeJSONResponse(w, statusCode, data)

			if w.Code != statusCode {
				t.Errorf("Expected status %d, got %d", statusCode, w.Code)
			}
		}
	})
}

func TestWriteErrorResponse(t *testing.T) {
	w := httptest.NewRecorder()

	writeErrorResponse(w, http.StatusBadRequest, "test error")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response Response
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "error" {
		t.Errorf("Expected status 'error', got '%s'", response.Status)
	}

	if response.Error != "test error" {
		t.Errorf("Expected error 'test error', got '%s'", response.Error)
	}
}
