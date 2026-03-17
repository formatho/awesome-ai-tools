// Package handlers provides HTTP request handlers for the Agent Orchestrator.
package handlers

import (
	"encoding/json"
	"net/http"
)

// jsonResponse writes a JSON response to the HTTP response writer.
func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
	}
}

// writeError writes an error response with the given status code and message.
func writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   message,
		"status":  http.StatusText(statusCode),
	})
}
