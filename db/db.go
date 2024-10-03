package db

import (
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

// NewMySQLStorage initializes a new MySQL database connection
func NewMySQLStorage(cfg mysql.Config) (*sql.DB, error) {
	// Create the DSN (Data Source Name) for MySQL connection
	dsn := cfg.FormatDSN()

	// Initialize the database connection (this doesn't actually establish a connection yet)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create MySQL connection: %w", err)
	}

	// // Configure the connection pool settings (optional, but recommended)
	// db.SetMaxOpenConns(25) // Maximum number of open connections to the database
	// db.SetMaxIdleConns(25) // Maximum number of idle connections in the pool
	// db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection by pinging the database
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL database: %w", err)
	}

	// Return the database connection
	return db, nil
}
