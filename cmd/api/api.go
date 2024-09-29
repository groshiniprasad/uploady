package api

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/groshiniprasad/uploady/services/user"
)

type APIServer struct {
	addr       string
	db         *sql.DB
	httpServer *http.Server
}

// NewAPIServer creates a new instance of APIServer
func NewAPIServer(addr string, db *sql.DB) *APIServer {
	return &APIServer{
		addr: addr,
		db:   db,
	}
}

// Run starts the server and listens on the provided address
func (s *APIServer) Run() error {
	// Create the router
	router := mux.NewRouter()

	// Create a subrouter for API versioning
	subrouter := router.PathPrefix("/api/v1").Subrouter()

	// Setup user routes
	userStore := user.NewStore(s.db)
	userHandler := user.NewHandler(userStore)
	userHandler.RegisterRoutes(subrouter)

	// Initialize the HTTP server
	s.httpServer = &http.Server{
		Addr:    s.addr,
		Handler: router,
	}

	log.Println("Listening on", s.addr)

	// Start the HTTP server
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

// Shutdown gracefully shuts down the server with a timeout
func (s *APIServer) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server...")
	return s.httpServer.Shutdown(ctx)
}
