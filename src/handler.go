package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
)

func getSMSHandler(w http.ResponseWriter, r *http.Request) {
	password := os.Getenv("PASSWORD")
	endpoint := os.Getenv("ENDPOINT")
	if password == "" || endpoint == "" {
		http.Error(w, "Server configuration error", http.StatusInternalServerError)
		log.Println("Environment variables PASSWORD or ENDPOINT not set")
		return
	}

	// Parse and validate GET parameters
	page, err := parseQueryParam(r, "page", 0, 0, 100) // Default: 0, Range: [0, 100]
	if err != nil {
		http.Error(w, "Invalid 'page' parameter", http.StatusBadRequest)
		return
	}

	perPage, err := parseQueryParam(r, "perPage", 500, 1, 1000) // Default: 500, Range: [1, 1000]
	if err != nil {
		http.Error(w, "Invalid 'perPage' parameter", http.StatusBadRequest)
		return
	}

	memStore, err := parseQueryParam(r, "memStore", 1, 0, 2) // Default: 1, Range: [0, 2]
	if err != nil {
		http.Error(w, "Invalid 'memStore' parameter", http.StatusBadRequest)
		return
	}

	tag, err := parseQueryParam(r, "tag", 10, 0, 10) // Default: 10, Range: [0, 10]
	if err != nil {
		http.Error(w, "Invalid 'tag' parameter", http.StatusBadRequest)
		return
	}

	zte, err := NewZTEConnector(endpoint)
	if err != nil {
		http.Error(w, "Failed to initialize connector", http.StatusInternalServerError)
		log.Printf("Connector initialization error: %v", err)
		return
	}

	if err := zte.Login(password); err != nil {
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		log.Printf("Login error: %v", err)
		return
	}

	smsList, err := zte.GetSMS(page, perPage, memStore, tag)
	if err != nil {
		http.Error(w, "Failed to retrieve SMS messages", http.StatusServiceUnavailable)
		log.Printf("GetSMS error: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(smsList); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		log.Printf("JSON encoding error: %v", err)
		return
	}
}

// Helper function to parse and validate query parameters
func parseQueryParam(r *http.Request, name string, defaultValue, min, max int) (int, error) {
	values := r.URL.Query()
	param := values.Get(name)
	if param == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(param)
	if err != nil || value < min || value > max {
		return 0, err
	}

	return value, nil
}
