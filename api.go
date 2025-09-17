package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// APIServer represents the API server with its configuration and storage
type APIServer struct {
	listenAddr string
	store    Storage
}

// NewAPIServer creates a new instance of APIServer with the given address and storage
func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}


func (s *APIServer) Run() {
	router := mux.NewRouter()
	router.HandleFunc("/healthz" , makeHTTPHandleFunc(s.handleProcessEmails))


	log.Println("JSON API server listening on " + s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)

}

func (s *APIServer) handleProcessEmails(w http.ResponseWriter, r *http.Request) error {
	log.Println("Processing emails...")
	count , err := s.StartProcessing()
	if err != nil{
		return err
	}
	resp:= map[string]interface{}{
		"status": "emails processed",
		"count": count,
	}
	return writeJSON(w, http.StatusOK, resp)

}

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

