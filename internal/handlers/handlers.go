package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Internal server error: failed to encode response", http.StatusInternalServerError)
	}
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	response := Response{
		Status: "error",
		Error:  message,
	}
	writeJSONResponse(w, statusCode, response)
}

func Health(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Status:  "ok",
		Message: "OpenBPL is running",
		Data: map[string]interface{}{
			"timestamp": time.Now().UTC(),
			"uptime":    "calculating...",
		},
	}

	writeJSONResponse(w, http.StatusOK, response)
}

func Status(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Status:  "ok",
		Message: "OpenBPL Status",
		Data: map[string]interface{}{
			"service":     "OpenBPL",
			"version":     "0.1.0",
			"environment": "development",
			"timestamp":   time.Now().UTC(),
			"endpoints": map[string]string{
				"health": "/health",
				"status": "/api/v1/status",
			},
		},
	}

	writeJSONResponse(w, http.StatusOK, response)
}

func Home(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Status:  "ok",
		Message: "Welcome to OpenBPL - Open Brand Protection Library",
		Data: map[string]interface{}{
			"description":   "An open-source framework for monitoring, detecting, and acting against brand infringements",
			"api_version":   "v1",
			"documentation": "/docs",
		},
	}

	writeJSONResponse(w, http.StatusOK, response)
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Status: "error",
		Error:  fmt.Sprintf("Endpoint not found: %s %s", r.Method, r.URL.Path),
	}

	writeJSONResponse(w, http.StatusNotFound, response)
}

func MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Status: "error",
		Error:  fmt.Sprintf("Method %s not allowed for %s", r.Method, r.URL.Path),
	}

	writeJSONResponse(w, http.StatusMethodNotAllowed, response)
}
