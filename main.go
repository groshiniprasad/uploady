package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/groshiniprasad/uploady/receipts"
	"github.com/groshiniprasad/uploady/utils"

	"github.com/lpernett/godotenv"
	"github.com/rs/cors"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file: ", err)
	}

	utils.CreateUploadsDir()

	r := mux.NewRouter()
	frontendUrl := os.Getenv("FRONTEND_URL")

	r.HandleFunc("/receipts", receipts.UploadReceiptImage).Methods("POST")
	r.HandleFunc("/receipts/{id}", receipts.GetReceiptImage).Methods("GET")

	// Add other routes here

	c := cors.New(cors.Options{
		AllowedOrigins: []string{frontendUrl},
		AllowedMethods: []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
		Debug:          true, // Enable debugging for development
	})

	// Wrap the router with the CORS handler
	handler := c.Handler(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server starting on port", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
