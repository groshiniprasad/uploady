package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type APIServer struct {
	addr string
	db   *sql.DB
}

func NewAPIServer(addr string, db *sql.DB) *APIServer {
	return &APIServer{
		addr: addr,
		db:   db,
	}
}

func (s *APIServer) Run() error {
	router := mux.NewRouter()
	// Serve static files
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("static")))

	log.Println("Listening on", s.addr)

	return http.ListenAndServe(s.addr, router)
}
