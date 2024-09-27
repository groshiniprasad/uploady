package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/groshiniprasad/uploady/receipts"
	"github.com/lpernett/godotenv"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/upload", receipts.UploadReceiptImage).Methods("POST")

	// load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("PORT")

	log.Println("Server starting on port ", port)
	log.Fatal(http.ListenAndServe(port, r))
}
