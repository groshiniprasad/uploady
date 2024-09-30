package receipt

import (
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/groshiniprasad/uploady/services/auth"
	"github.com/groshiniprasad/uploady/types"
	"github.com/groshiniprasad/uploady/utils"
)

type Handler struct {
	store     types.ReceiptStore
	userStore types.UserStore
}

func NewHandler(store types.ReceiptStore, userStore types.UserStore) *Handler {
	return &Handler{store: store, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	// Routers to get all receipts of a user

	router.HandleFunc("/receipts/upload", auth.WithJWTAuth(h.handleCreateReceipt, h.userStore)).Methods(http.MethodPost)
	router.HandleFunc("/receipts/{id}", auth.WithJWTAuth(h.handleGetResizedReceiptsV2, h.userStore)).Methods(http.MethodGet)

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

	_, err = h.store.CreateReceipt(receipt)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// Respond with success
	fmt.Fprintf(w, "Receipt uploaded successfully: %+v\n", receipt)

}

func (h *Handler) handleGetResizedReceiptsV2(w http.ResponseWriter, r *http.Request) {
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

	receipt, err := h.store.GetReceiptByID(receiptID, userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// Open the file
	file, err := os.Open(receipt.ImagePath)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	defer file.Close()

	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error decoding image: %v", err))
		return
	}

	width, height := utils.GetWidthHeightFromQuery(r)

	// Resize the image
	resizedImg := utils.ResizeImage(img, width, height)

	// Set the content type
	w.Header().Set("Content-Type", "image/jpeg")

	// Encode and write the resized image to the response writer
	err = jpeg.Encode(w, resizedImg, nil)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error encoding image: %v", err))
		return
	}
}

//func (h *Handler) handleGetReceipts(w http.ResponseWriter, r *http.Request) {}
