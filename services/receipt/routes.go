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
	filePath := fmt.Sprintf("./uploads/%s", filename)
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

	receiptId, err := h.store.CreateReceipt(receipt)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	cacheKey := fmt.Sprintf("user_%d_receipt_%d", userID, receiptId)
	h.cache.Set(cacheKey, filePath, 5*time.Minute)
	// Respond with success
	fmt.Fprintf(w, "Receipt uploaded successfully: %+v\n", receipt)

}

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

	// Prepare payload for the resizing task
	payload := types.ResizeTaskPayload{
		ImagePath: imagePath,
		Width:     width,
		Height:    height,
		Response:  w,
	}
	// Submit the task to the worker pool and wait for it to complete
	err = h.workerPool.AddTask(func() error {
		return ResizeAndServeImageV2(payload, w)
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
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
	resizedImg := utils.ResizeImage(img, payload.Width, payload.Height)

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
