package utils

import (
	"fmt"
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
	return uuid.New().String() + originalFilename
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

func ResizeImage(img image.Image, width, height int) (image.Image, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("Dimensions Invalid") // Return original image if dimensions are invalid
	}
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
	return dst, nil
}

// CropExactPart crops a specific rectangular part of the image based on the given coordinates (x, y, width, height)
func CropExactPart(img image.Image, x, y, width, height int) (image.Image, error) {
	// Get the bounds of the original image
	bounds := img.Bounds()

	// Ensure that the coordinates and dimensions are within the bounds of the original image
	if x < 0 || y < 0 || x+width > bounds.Dx() || y+height > bounds.Dy() {
		return nil, fmt.Errorf("crop area exceeds image bounds")
	}

	// Define the cropping rectangle based on the coordinates and dimensions
	cropRect := image.Rect(x, y, x+width, y+height)

	// Create a new image with the dimensions of the cropped area
	croppedImg := image.NewRGBA(cropRect)

	draw.Draw(croppedImg, croppedImg.Bounds(), img, image.Point{x, y}, draw.Src)

	return croppedImg, nil
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

// Helper to get X and Y from query parameters
func GetXYFromQuery(r *http.Request) (*int, *int) {
	xStr := r.URL.Query().Get("x")
	yStr := r.URL.Query().Get("y")

	var x, y *int

	if xVal, err := strconv.Atoi(xStr); err == nil && xVal > 0 {
		x = &xVal // Valid x
	}

	if yVal, err := strconv.Atoi(yStr); err == nil && yVal > 0 {
		y = &yVal // Valid y
	}

	return x, y
}
