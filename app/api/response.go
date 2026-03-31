package api

import (
	"encoding/json"
	"net/http"
)

// OKResponse writes a 200 OK JSON response with the provided data.
// Sets Content-Type: application/json header and marshals data directly to JSON.
// The response body contains the data as-is with no wrapper key.
func OKResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

// ErrorResponse writes a JSON error response with the given HTTP status code.
// Sets Content-Type: application/json header.
// The response body is formatted as {"error": "<message>"}.
func ErrorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
