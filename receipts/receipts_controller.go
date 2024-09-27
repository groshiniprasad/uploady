package receipts

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/groshiniprasad/uploady/utils"
	"golang.org/x/image/draw"
)

func UploadReceiptImage(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("image")
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid file upload")
		return
	}
	defer file.Close()

	isValid, errMsg := utils.IsValidImageFile(file, header)
	if !isValid {
		utils.RespondWithError(w, http.StatusBadRequest, errMsg)
		return
	}

	filename := utils.GenerateUniqueFilename(header.Filename)
	filepath := filepath.Join("uploads", filename)

	out, err := os.Create(filepath)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create file")
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to save file")
		return
	}

	receipt := Receipt{
		ID:       filename,
		Filename: header.Filename,
	}
	SaveReceipt(receipt)

	utils.RespondWithJSON(w, http.StatusCreated, receipt)
}

func GetReceiptImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	filepath := filepath.Join("uploads", id)

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		utils.RespondWithError(w, http.StatusNotFound, "Image not found")
		return
	}

	file, err := os.Open(filepath)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to open file")
		return
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to decode image")
		return
	}

	width, _ := strconv.Atoi(r.URL.Query().Get("width"))
	height, _ := strconv.Atoi(r.URL.Query().Get("height"))

	if width > 0 || height > 0 {
		// Calculate new dimensions while maintaining aspect ratio
		origBounds := img.Bounds()
		origWidth := origBounds.Dx()
		origHeight := origBounds.Dy()

		if width == 0 {
			width = origWidth * height / origHeight
		} else if height == 0 {
			height = origHeight * width / origWidth
		}

		// Create a new image with the calculated dimensions
		newImg := image.NewRGBA(image.Rect(0, 0, width, height))

		// Use ApproxBiLinear to resize the image
		draw.ApproxBiLinear.Scale(newImg, newImg.Bounds(), img, img.Bounds(), draw.Over, nil)

		img = newImg
	}

	w.Header().Set("Content-Type", fmt.Sprintf("image/%s", format))

	switch format {
	case "jpeg":
		jpeg.Encode(w, img, nil)
	case "png":
		png.Encode(w, img)
	default:
		utils.RespondWithError(w, http.StatusInternalServerError, "Unsupported image format")
	}
}
