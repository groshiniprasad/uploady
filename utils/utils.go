package utils

import (
	"image"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/image/draw"

	"github.com/google/uuid"
)

const MaxFileSize = 10 << 20 // 10 MB

func IsValidImageFile(file multipart.File, header *multipart.FileHeader) (bool, string) {
	// Check file size
	if header.Size > MaxFileSize {
		return false, "File size exceeds the limit (10 MB), please upload a smaller file"
	}

	// Check file extension
	filename := header.Filename
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return false, "Only JPG and PNG files are allowed, please upload a valid image file"
	}

	// Check file content type
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil {
		return false, "Error reading file"
	}
	file.Seek(0, io.SeekStart) // Reset file pointer to beginning

	contentType := http.DetectContentType(buffer)
	if contentType != "image/jpeg" && contentType != "image/png" {
		return false, "File is not a valid image, please upload a valid image file"
	}

	return true, ""
}

func GenerateUniqueFilename(originalFilename string) string {
	// Generate a unique filename using UUID
	extension := filepath.Ext(originalFilename)
	return uuid.New().String() + extension
}

func CreateUploadsDir() {
	dir := "./uploads"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.Mkdir(dir, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func ResizeImage(img image.Image, width, height int) image.Image {
	if width <= 0 || height <= 0 {
		return img // Return original image if dimensions are invalid
	}
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
	return dst
}

// Helper to get width and height from query parameters
func GetWidthHeightFromQuery(r *http.Request) (int, int) {
	widthStr := r.URL.Query().Get("width")
	heightStr := r.URL.Query().Get("height")

	width, err := strconv.Atoi(widthStr)
	if err != nil || width <= 0 {
		width = 100 // default width
	}

	height, err := strconv.Atoi(heightStr)
	if err != nil || height <= 0 {
		height = 100 // default height
	}

	return width, height
}
