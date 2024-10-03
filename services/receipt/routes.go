package receipt

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/groshiniprasad/uploady/cache"
	"github.com/groshiniprasad/uploady/services/auth"
	"github.com/groshiniprasad/uploady/types"
	"github.com/groshiniprasad/uploady/utils"
	"github.com/groshiniprasad/uploady/worker"
)

type Handler struct {
	store      types.ReceiptStore
	userStore  types.UserStore
	cache      *cache.Cache
	workerPool *worker.WorkerPool
}

func NewHandler(store types.ReceiptStore, userStore types.UserStore, cacheStore *cache.Cache, workerPool *worker.WorkerPool) *Handler {
	return &Handler{store: store, userStore: userStore, cache: cacheStore, workerPool: workerPool}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	// Routers to get all receipts of a user

	router.HandleFunc("/receipts/upload", auth.WithJWTAuth(h.handleCreateReceipt, h.userStore)).Methods(http.MethodPost)
	router.HandleFunc("/receipts/{id}", auth.WithJWTAuth(h.handleGetResizedReceiptsV5, h.userStore)).Methods(http.MethodGet)

}

func (h *Handler) handleCreateReceipt(w http.ResponseWriter, r *http.Request) {

	userID := auth.GetUserIDFromContext(r.Context())

	// Parse the multipart form, limit file size to 10MB
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Extract the metadata from the form
	name := r.FormValue("name")
	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	dateStr := r.FormValue("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	// Get the file from the form
	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error uploading file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate the file using the isValidFile function
	isValid, validationMsg := utils.IsValidImageFile(file, fileHeader)
	if !isValid {
		http.Error(w, validationMsg, http.StatusBadRequest)
		return
	}

	// Save the image file to disk (or cloud storage)
	filename := utils.GenerateUniqueFilename(fileHeader.Filename)
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}
	filePath := fmt.Sprintf("%s/uploads/%s", cwd, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to save file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy the file data to the destination
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Error saving file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// date is now of type time.Timeeipt object (this could be inserted into a database)
	receipt := types.Receipt{
		UserID:    userID,
		Name:      name,
		Amount:    amount,
		Date:      date,
		ImagePath: filePath, // Save the path where the image is stored
	}

	_, err = h.store.CreateReceipt(receipt)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// cacheKey := fmt.Sprintf("user_%d_receipt_%d", userID, receiptId)
	// h.cache.Set(cacheKey, filePath, 5*time.Minute)
	// Respond with success
	fmt.Fprintf(w, "Receipt uploaded successfully: %+v\n", receipt)

}

// This handler function reads the receipt image from disk, resizes it, and sends the resized image in the response  && later I have tested croping the image functionality f
func (h *Handler) handleGetResizedReceiptsV5(w http.ResponseWriter, r *http.Request) {
	startNow := time.Now()

	vars := mux.Vars(r)
	userID := auth.GetUserIDFromContext(r.Context())

	str, ok := vars["id"]
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing receipt ID"))
		return
	}

	receiptID, err := strconv.Atoi(str)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid receipt ID"))
		return
	}

	// Log cache size for debugging
	log.Printf("Cache size: %d items", h.cache.Size())

	// Cache key
	cacheKey := fmt.Sprintf("user_%d_receipt_%d", userID, receiptID)
	var imagePath string

	// Check if the value is in the cache
	if filePathValue, found := h.cache.Get(cacheKey); found {
		// Cache hit
		log.Printf("Cache hit: %s", cacheKey)
		imagePath, ok = filePathValue.(string)
		if !ok {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("cached value is not a string"))
			return
		}
	} else {
		// Cache miss
		log.Printf("Cache miss: %s", cacheKey)

		// Fetch the receipt (heavy task processing happens in worker)
		receipt, err := h.store.GetReceiptByID(receiptID, userID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		// Store the imagePath in the cache for future requests
		imagePath = receipt.ImagePath
		h.cache.Set(cacheKey, imagePath, 5*time.Minute)
	}

	// Get width and height parameters from the query string
	width, height := utils.GetWidthHeightFromQuery(r)

	x, y := utils.GetXYFromQuery(r)

	// Check if x and y are provided in the query, assuming -1 means not provided
	if x != nil && y != nil {
		// Prepare crop payload
		cropPayload := types.CropTaskPayload{
			ImagePath: imagePath,
			Width:     width,
			Height:    height,
			Response:  w,
			X:         *x,
			Y:         *y,
		}

		// Submit the crop task to the worker pool
		err := h.workerPool.AddTask(func() error {
			return CropAndServeImageV2(cropPayload, w)
		})

		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		fmt.Fprintf(w, "Image cropping task submitted")
	} else {
		// Prepare resize payload
		resizePayload := types.ResizeTaskPayload{
			ImagePath: imagePath,
			Width:     width,
			Height:    height,
			Response:  w,
		}

		// Submit the resize task to the worker pool
		err := h.workerPool.AddTask(func() error {
			return ResizeAndServeImageV2(resizePayload, w)
		})

		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		fmt.Fprintf(w, "Image resizing task submitted")
	}

	log.Printf("Image resizing completed in %v", time.Since(startNow))
}

// ResizeAndServeImage performs the actual image resizing and sends the response
func ResizeAndServeImageV2(payload types.ResizeTaskPayload, w http.ResponseWriter) error {
	// Open the file
	file, err := os.Open(payload.ImagePath)
	if err != nil {
		log.Printf("Failed to open image file: %v\n", err)
		return fmt.Errorf("failed to open image file")
	}
	defer file.Close()

	// Decode the image
	img, format, err := image.Decode(file)
	if err != nil {
		log.Printf("Error decoding image: %v\n", err)
		return fmt.Errorf("error decoding image")
	}

	// Resize the image

	resizedImg, _ := utils.ResizeImage(img, payload.Width, payload.Height)

	// Write the resized image to a buffer first
	buf := new(bytes.Buffer)

	// Set the appropriate content type based on the image format
	switch format {
	case "jpeg", "jpg":
		w.Header().Set("Content-Type", "image/jpeg")
		err = jpeg.Encode(buf, resizedImg, nil) // Encode the image to the buffer
	case "png":
		w.Header().Set("Content-Type", "image/png")
		err = png.Encode(buf, resizedImg) // PNG encoder, ensure you import "image/png"
	default:
		log.Printf("Unsupported image format: %s", format)
		return fmt.Errorf("unsupported image format")
	}

	if err != nil {
		log.Printf("Error encoding image: %v\n", err)
		return fmt.Errorf("error encoding image")
	}

	// Now write the buffer to the response writer
	_, err = buf.WriteTo(w) // Write the buffer content to the ResponseWriter
	if err != nil {
		log.Printf("Error writing image to response: %v\n", err)
		return fmt.Errorf("error writing image to response")
	}

	log.Printf("Image resized and served successfully")
	return nil
}

func CropAndServeImageV2(payload types.CropTaskPayload, w http.ResponseWriter) error {
	// Open the file
	file, err := os.Open(payload.ImagePath)
	if err != nil {
		log.Printf("Failed to open image file: %v\n", err)
		return fmt.Errorf("failed to open image file")
	}
	defer file.Close()

	// Decode the image
	img, format, err := image.Decode(file)
	if err != nil {
		log.Printf("Error decoding image: %v\n", err)
		return fmt.Errorf("error decoding image")
	}

	// Resize the image
	// resizedImg, _ := utils.ResizeImage(img, payload.Width, payload.Height)

	resizedImg, _ := utils.CropExactPart(img, payload.X, payload.Y, payload.Width, payload.Height)

	// Write the resized image to a buffer first
	buf := new(bytes.Buffer)

	// Set the appropriate content type based on the image format
	switch format {
	case "jpeg", "jpg":
		w.Header().Set("Content-Type", "image/jpeg")
		err = jpeg.Encode(buf, resizedImg, nil) // Encode the image to the buffer
	case "png":
		w.Header().Set("Content-Type", "image/png")
		err = png.Encode(buf, resizedImg) // PNG encoder, ensure you import "image/png"
	default:
		log.Printf("Unsupported image format: %s", format)
		return fmt.Errorf("unsupported image format")
	}

	if err != nil {
		log.Printf("Error encoding image: %v\n", err)
		return fmt.Errorf("error encoding image")
	}

	// Now write the buffer to the response writer
	_, err = buf.WriteTo(w) // Write the buffer content to the ResponseWriter
	if err != nil {
		log.Printf("Error writing image to response: %v\n", err)
		return fmt.Errorf("error writing image to response")
	}

	log.Printf("Image resized and served successfully")
	return nil
}

func (h *Handler) handleGetResizedReceiptsImgCaching(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := auth.GetUserIDFromContext(r.Context())
	receiptID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid receipt ID"))
		return
	}

	cacheKey := fmt.Sprintf("user_%d_receipt_%d", userID, receiptID)
	var img image.Imagée

	// Check if the image is already cached
	if cachedImg, found := h.cache.Get(cacheKey); found {
		img, ok := cachedImg.(image.Image)
		if ok {
			// Use the cached image, no need to read from disk
			log.Printf("Cache hit: %s", cacheKey)
			// Continue with resizing and sending the response
			resizeAndSendImage(img, w, r)
			return
		}
	}

	// Read image from disk only if not cached
	receipt, err := h.store.GetReceiptByID(receiptID, userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	img, err = h.readImageFromDisk(receipt.ImagePath)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// Cache the image after reading from disk
	h.cache.Set(cacheKey, img, 5*time.Minute)

	// Proceed with resizing and sending the response
	resizeAndSendImage(img, w, r)
}

func resizeAndSendImage(img image.Image, w http.ResponseWriter, r *http.Request) {
	// Extract width and height parameters from the query string
	width, height := utils.GetWidthHeightFromQuery(r)

	// Resize the image
	resizedImg, _ := utils.ResizeImage(img, int(width), int(height))

	// Buffer to hold the encoded image before writing to the response
	var buf bytes.Buffer

	// Determine the format of the image (JPEG, PNG, etc.)
	format := utils.GetFormatFromRequest(r)

	// Set the appropriate content type
	if format == "jpeg" {
		w.Header().Set("Content-Type", "image/jpeg")
		err := jpeg.Encode(&buf, resizedImg, nil)
		if err != nil {
			http.Error(w, "Error encoding image", http.StatusInternalServerError)
			return
		}
	} else if format == "png" {
		w.Header().Set("Content-Type", "image/png")
		err := png.Encode(&buf, resizedImg)
		if err != nil {
			http.Error(w, "Error encoding image", http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "Unsupported image format", http.StatusBadRequest)
		return
	}

	// Write the buffered image data to the response
	w.WriteHeader(http.StatusOK)
	_, err := buf.WriteTo(w)
	if err != nil {
		log.Printf("Error writing image to response: %v", err)
		http.Error(w, "Error writing image to response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) readImageFromDisk(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file: %v", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("error decoding image: %v", err)
	}
	return img, nil
}

//func (h *Handler) handleGetReceipts(w http.ResponseWriter, r *http.Request) {}
