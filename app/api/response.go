package api

import (
	"encoding/json"
	"log"
	"net/http"
)

// JSON writes a JSON response with the given HTTP status code and data.
// Sets Content-Type: application/json header and marshals data directly to JSON.
// This is a generic helper used by specific response functions.
// If encoding fails, the error is logged but does not panic.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log encoding error but don't panic; response headers already sent
		log.Printf("failed to encode JSON response: %v", err)
	}
}

// OKResponse writes a 200 OK JSON response with the provided data.
// Sets Content-Type: application/json header and marshals data directly to JSON.
// The response body contains the data as-is with no wrapper key.
func OKResponse(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, data)
}

// CreatedResponse writes a 201 Created JSON response with the provided data.
// Sets Content-Type: application/json header and marshals data directly to JSON.
// The response body contains the data as-is with no wrapper key.
func CreatedResponse(w http.ResponseWriter, data any) {
	JSON(w, http.StatusCreated, data)
}

// ErrorResponse writes a JSON error response with the given HTTP status code.
// Sets Content-Type: application/json header.
// The response body is formatted as {"error": "<message>"}.
func ErrorResponse(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{"error": message})
}
