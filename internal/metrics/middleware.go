// Package metrics provides HTTP middleware for collecting request metrics.
package metrics

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// HTTPMiddleware wraps an HTTP handler to collect metrics
type HTTPMiddleware struct {
	metrics *Metrics
}

// NewHTTPMiddleware creates a new HTTP metrics middleware
func NewHTTPMiddleware() *HTTPMiddleware {
	return &HTTPMiddleware{
		metrics: GetMetrics(),
	}
}

// Handler wraps an HTTP handler with metrics collection
func (m *HTTPMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Track in-flight requests
		m.metrics.IncHTTPRequestsInFlight()
		defer m.metrics.DecHTTPRequestsInFlight()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Call the next handler
		next.ServeHTTP(wrapped, r)

		// Record metrics
		duration := time.Since(start)
		path := normalizePath(r.URL.Path)
		status := strconv.Itoa(wrapped.statusCode)

		m.metrics.RecordHTTPRequest(r.Method, path, status, duration)
	})
}

// HandlerFunc wraps an HTTP handler function with metrics collection
func (m *HTTPMiddleware) HandlerFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.Handler(next).ServeHTTP(w, r)
	}
}

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// WriteHeader captures the status code
func (w *responseWriter) WriteHeader(statusCode int) {
	if !w.written {
		w.statusCode = statusCode
		w.written = true
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the status code if not already set
func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.statusCode = http.StatusOK
		w.written = true
	}
	return w.ResponseWriter.Write(b)
}

// normalizePath normalizes URL paths for metrics labels
// This prevents high cardinality by replacing IDs with placeholders
func normalizePath(path string) string {
	parts := strings.Split(path, "/")
	normalized := make([]string, len(parts))

	for i, part := range parts {
		// Check if this looks like a UUID or numeric ID
		if isUUID(part) || isNumericID(part) {
			normalized[i] = ":id"
		} else {
			normalized[i] = part
		}
	}

	return strings.Join(normalized, "/")
}

// isUUID checks if a string looks like a UUID
func isUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	// Check for UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
		} else {
			if !isHexChar(byte(c)) {
				return false
			}
		}
	}
	return true
}

// isHexChar checks if a byte is a hexadecimal character
func isHexChar(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

// isNumericID checks if a string is a numeric ID
func isNumericID(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
