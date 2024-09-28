package api

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type APIServer struct {
	addr       string
	db         *sql.DB
	httpServer *http.Server
}

func NewAPIServer(addr string, db *sql.DB) *APIServer {
	server := &APIServer{
		addr: addr,
		db:   db,
	}
	router := mux.NewRouter()

	// Serve static files
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("static")))

	server.httpServer = &http.Server{
		Addr:    addr,
		Handler: router,
	}

	return server
}

func (s *APIServer) Run() error {
	log.Println("Listening on", s.addr)

	// Start the server
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
