package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

var Validate = validator.New()

func GetTokenFromRequest(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "" // No Authorization header present
	}

	// Split the Authorization header into parts
	parts := strings.Split(authHeader, " ")
	if len(parts) == 2 && parts[0] == "Bearer" {
		return parts[1] // Return the token part
	}

	return "" // Invalid format or no token found
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	// Set the content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Set the status code
	w.WriteHeader(status)

	// Encode the data to JSON and write it to the response
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
	}
}

func WriteError(w http.ResponseWriter, status int, err error) {
	// Check if the error is nil to avoid calling err.Error() on a nil reference
	if err == nil {
		err = fmt.Errorf("unknown error")
	}

	// Write the error message as a JSON response
	WriteJSON(w, status, map[string]string{"error": err.Error()})
}

func ParseJSON(r *http.Request, v any) error {
	if r.Body == nil {
		return fmt.Errorf("missing request body")
	}

	return json.NewDecoder(r.Body).Decode(v)
}
