package main

import (
	"context"
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
	"github.com/groshiniprasad/uploady/utils"
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
	database, err := db.NewMySQLStorage(cfg)
	if err != nil {
		log.Fatalf("Could not connect to MySQL: %v", err)
	}
	defer database.Close()

	fmt.Println("Database successfully connected!")

	// Initialize DB connection
	// Setup API server
	server := api.NewAPIServer(fmt.Sprintf(":%s", configs.Envs.Port), database)

	// Start server in a goroutine to allow for graceful shutdown
	go func() {
		if err := server.Run(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to run API server: ", err)
		}
	}()

	utils.CreateUploadsDir()

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
