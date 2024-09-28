package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/groshiniprasad/uploady/cmd/api"
	"github.com/groshiniprasad/uploady/configs"
	"github.com/groshiniprasad/uploady/db"
)

func main() {
	// MySQL connection configuration
	cfg := mysql.Config{
		User:                 configs.Envs.DBUser,
		Passwd:               configs.Envs.DBPassword,
		Addr:                 configs.Envs.DBAddress,
		DBName:               configs.Envs.DBName,
		Net:                  "tcp",
		AllowNativePasswords: true,
		ParseTime:            true,
	}

	// Initialize DB
	db, err := db.NewMySQLStorage(cfg)
	if err != nil {
		log.Fatal("Failed to initialize database: ", err)
	}
	defer db.Close()

	// Initialize DB connection
	initStorage(db)

	// Setup API server
	server := api.NewAPIServer(fmt.Sprintf(":%s", configs.Envs.Port), db)

	// Start server in a goroutine to allow for graceful shutdown
	go func() {
		if err := server.Run(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to run API server: ", err)
		}
	}()

	// Create a channel to listen for interrupt or terminate signals
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal in shutdownChan
	<-shutdownChan
	log.Println("Shutdown signal received, shutting down gracefully...")

	// Create a deadline for the shutdown (e.g., 5 seconds)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server gracefully, ensuring no new connections are accepted
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server gracefully stopped.")
}

// initStorage pings the DB and ensures it's reachable
func initStorage(db *sql.DB) {
	err := db.Ping()
	if err != nil {
		log.Fatal("Unable to connect to the database: ", err)
	}

	log.Println("DB: Successfully connected!")
}
