package main

import (
	"encoding/json"
	"net/http"
)

// apiFunc is a custom type that represents an API handler function
type apiFunc func(w http.ResponseWriter, r *http.Request) error

// apiError represents a structured error response for the API
type apiError struct {
	Error   error  `json:"error"`
	Message string `json:"message"`
}

// writeJSON writes a JSON response with the given status code and value
func writeJSON( w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// makeHTTPHandleFunc converts an apiFunc into an http.HandlerFunc
func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w,r);err !=nil{
			writeJSON(w, http.StatusBadRequest, apiError{Error: err, Message: err.Error()} )
		}
	}
}

