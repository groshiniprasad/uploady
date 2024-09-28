package main

import (
	"log"
	"os"
	"path/filepath"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	mysqlMigrate "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/groshiniprasad/uploady/configs"
)

func main() {
	// MySQL connection config
	cfg := mysqlDriver.Config{
		User:                 configs.Envs.DBUser,
		Passwd:               configs.Envs.DBPassword,
		Addr:                 configs.Envs.DBAddress,
		DBName:               configs.Envs.DBName,
		Net:                  "tcp",
		AllowNativePasswords: true,
		ParseTime:            true,
	}

	// Open the database connection
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Get the current working directory (pwd)
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	// Construct the path to the migrations folder dynamically
	migrationsPath := filepath.Join(pwd, "cmd", "migrate", "migrations")

	// Create the migration driver instance
	driver, err := mysqlMigrate.WithInstance(db, &mysqlMigrate.Config{})
	if err != nil {
		log.Fatalf("Failed to create migration driver: %v", err)
	}

	// Initialize the migration with the dynamically set path
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath, // Absolute path to migrations
		"mysql",
		driver,
	)
	if err != nil {
		log.Fatalf("Failed to initialize migration: %v", err)
	}

	// Proceed with migration commands (up or down)
	cmd := os.Args[len(os.Args)-1]
	if cmd == "up" {
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("Migrations applied successfully.")
	} else if cmd == "down" {
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Rollback failed: %v", err)
		}
		log.Println("Rollback successful.")
	} else {
		log.Fatalf("Invalid command. Use 'up' or 'down'.")
	}
}
