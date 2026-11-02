// Package response provides HTTP response formatting and utility helpers.
package response

import (
	"encoding/json"
	"net/http"
	"strings"
)

// JSON sends a JSON response with status code and standard security headers.
func JSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// Error writes a standardized JSON error message.
func Error(w http.ResponseWriter, code int, msg string) {
	JSON(w, code, map[string]string{"error": msg})
}

// DBError returns a 503 Service Unavailable when database operations fail.
func DBError(w http.ResponseWriter, err error) {
	JSON(w, http.StatusServiceUnavailable, map[string]string{
		"error":  "database unavailable",
		"detail": err.Error(),
	})
}

// ResolveDeviceID extracts device_id from URL query params, falling back to defaultID.
func ResolveDeviceID(r *http.Request, defaultID string) string {
	dev := r.URL.Query().Get("device_id")
	if dev == "" {
		dev = defaultID
	}
	return strings.TrimSpace(dev)
}

