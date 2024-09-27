package receipts

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/groshiniprasad/uploady/utils"
)

func UploadReceiptImage(w http.ResponseWriter, r *http.Request) {
	// Set the max memory for parsing the multipart form
	r.ParseMultipartForm(utils.MaxFileSize)

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate the image file
	isValid, errMsg := utils.ValidateImageFile(file, header)
	if !isValid {
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	// Generate a unique filename
	filename := utils.GenerateUniqueFilename(header.Filename)
	filepath := filepath.Join("uploads", filename)

	out, err := os.Create(filepath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "File uploaded successfully: %s", filename)
}
