package middleware

import (
	"log"
	"net/http"
	"time"
)

type Middleware func(http.Handler) http.Handler

// responseWriter wraps http.ResponseWriter to capture status code
// This is called "embedding" - we embed http.ResponseWriter to get all its methods
// Then we override specific methods to add our own behavior
type responseWriter struct {
	http.ResponseWriter     // Embedded field - gives us all ResponseWriter methods
	statusCode          int // Our additional field to track status code
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code                // Remember the status code
	rw.ResponseWriter.WriteHeader(code) // Call the original WriteHeader
}

// This chains middlewares onto a given handler, for example if you
// want A + B + C on handler h you essentially do  A(B(C(h)))
// we go in reverse to get the right execution order
func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
	for idx := len(middlewares) - 1; idx >= 0; idx-- {
		handler = middlewares[idx](handler)
	}
	return handler
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrappedWriter := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		next.ServeHTTP(wrappedWriter, r)
		duration := time.Since(start)
		log.Printf(
			"%s %s %d %v %s",         // Format string
			r.Method,                 // HTTP method (GET, POST, etc.)
			r.URL.Path,               // Request path (/api/users)
			wrappedWriter.statusCode, // HTTP status code (200, 404, etc.)
			duration,                 // How long the request took
			r.RemoteAddr,             // Client IP address
		)
	})
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")                                // Allow all origins
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS") // Allowed HTTP methods
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")     // Allowed headers
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK) // Return 200 OK
			return                       // Don't call the next handler - this is just a preflight check
		}

		// For non-OPTIONS requests, continue to the next handler
		next.ServeHTTP(w, r)
	})
}

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			// recover() captures any panic that occurred
			// If no panic occurred, recover() returns nil
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err) // Log the panic
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		// Call the next handler
		// If this panics, the defer function above will catch it
		next.ServeHTTP(w, r)
	})
}
