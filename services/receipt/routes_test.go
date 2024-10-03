package receipt

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/groshiniprasad/uploady/cache"
	"github.com/groshiniprasad/uploady/types"
	"github.com/groshiniprasad/uploady/worker"
	"github.com/stretchr/testify/assert"
)

func TestReceiptServiceHandlers(t *testing.T) {

	receiptStore := &mockReceiptStore{}
	userStore := &mockUserStore{}
	cache := cache.NewCache()
	workerPool := worker.NewWorkerPool(5, 10) // Assuming worker.WorkerPool is a struct, otherwise use a proper mock implementation
	handler := NewHandler(receiptStore, userStore, cache, workerPool)

	t.Run("should fail if the receipt ID is not a number", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/receipt/abc", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		router := mux.NewRouter()

		router.HandleFunc("/receipt/{receiptID}", handler.handleGetResizedReceiptsV5).Methods(http.MethodGet)

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})
	t.Run("should fail if wrong image format is added", func(t *testing.T) {
		// Create a multipart form request with a non-image file
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)

		// Mock form data
		writer.WriteField("name", "Test Receipt")
		writer.WriteField("amount", "100.50")
		writer.WriteField("date", "2023-01-01")

		// Add a non-image file (e.g., a text file)
		fileWriter, err := writer.CreateFormFile("image", "test.txt")
		assert.NoError(t, err)
		_, err = fileWriter.Write([]byte("This is not an image file"))
		assert.NoError(t, err)

		writer.Close()

		// Create a new POST request with the invalid image file
		req := httptest.NewRequest(http.MethodPost, "/receipts/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		// Record the response
		rr := httptest.NewRecorder()

		// Set up the router and handler
		router := mux.NewRouter()
		router.HandleFunc("/receipts/upload", handler.handleCreateReceipt).Methods(http.MethodPost)

		// Serve the HTTP request
		router.ServeHTTP(rr, req)

		// Assert that the server returns a 400 Bad Request status code due to invalid file format
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "please upload a valid image file")
	})

	t.Run("should successfully upload a valid image", func(t *testing.T) {
		// Create a buffer to hold the multipart form data
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)

		// Add valid form fields (metadata)
		writer.WriteField("name", "Valid Receipt")
		writer.WriteField("amount", "123.45")
		writer.WriteField("date", "2023-10-01")

		// Create a valid image file in the form
		fileWriter, err := writer.CreateFormFile("image", "test.jpg")
		assert.NoError(t, err)

		// Write a simple image to the file (a 100x100 pixel JPEG image)
		img := image.NewRGBA(image.Rect(0, 0, 100, 100))
		err = jpeg.Encode(fileWriter, img, nil)
		assert.NoError(t, err)

		// Close the writer to finalize the multipart form data
		writer.Close()

		// Create a new POST request to simulate the upload
		req := httptest.NewRequest(http.MethodPost, "/receipts/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		// Record the response
		rr := httptest.NewRecorder()

		// Set up the router and handler
		router := mux.NewRouter()
		router.HandleFunc("/receipts/upload", handler.handleCreateReceipt).Methods(http.MethodPost)

		// Serve the HTTP request
		router.ServeHTTP(rr, req)

		// Assert that the server returns a 200 OK status code
		assert.Equal(t, http.StatusOK, rr.Code)

		// Assert that the response contains a success message
		assert.Contains(t, rr.Body.String(), "test.jpg")

		// Check if the upload directory exists
		uploadDir := "./uploads"
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			err := os.Mkdir(uploadDir, os.ModePerm)
			if err != nil {
				t.Fatal(err)
				return
			}
		}
		// Check if the file has been created in the upload directory
		expectedFile := fmt.Sprintf("%s/test.jpg", uploadDir)
		_, err = os.Stat(expectedFile)
		assert.NoError(t, err, "Uploaded file should exist in the uploads directory")
	})

	t.Run("should handle get receipt by ID", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/receipt/1?width=50&height=50", nil)
		// Create the mock stores and services
		receiptStore := &mockReceiptStore{}
		userStore := &mockUserStore{}
		workerPool := worker.NewWorkerPool(3, 100)

		// Initialize the handler
		handler := NewHandler(receiptStore, userStore, cache, workerPool)

		// Create the request and response recorder
		req, err = http.NewRequest(http.MethodGet, "/receipt/1?width=200&height=200", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		router := mux.NewRouter()

		// Register the route with the handler
		router.HandleFunc("/receipt/{receiptID}", handler.handleGetResizedReceiptsV5).Methods(http.MethodGet)

		// Serve the request
		router.ServeHTTP(rr, req)
		assert.NoError(t, err)
		// Assert the status code
		if rr.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
		}
		// Optionally, check dimensions of the resized image

	})

}

type mockUserStore struct{}
type mockReceiptStore struct{}

// CreateReceipt implements types.ReceiptStore.
func (m *mockReceiptStore) CreateReceipt(types.Receipt) (int, error) {
	return 1, nil
}

// GetReceiptByID implements types.ReceiptStore.
func (m *mockReceiptStore) GetReceiptByID(receiptId int, userId int) (*types.Receipt, error) {
	return &types.Receipt{
		ID:        receiptId,
		UserID:    userId,
		Name:      "Test Receipt",
		Amount:    123.45,
		ImagePath: "./uploads/test.jpg", // Path to a test image file
	}, nil
}

// GetReceiptByName implements types.ReceiptStore.
func (m *mockReceiptStore) GetReceiptByName(name string, userId int) (*types.User, error) {
	panic("unimplemented")
}

type mockCache struct{}
type mockWorkerPool struct{}

// CreateUser implements types.UserStore.
func (m *mockUserStore) CreateUser(types.User) error {
	panic("unimplemented")
}

// GetUserByEmail implements types.UserStore.
func (m *mockUserStore) GetUserByEmail(email string) (*types.User, error) {
	panic("unimplemented")
}

// GetUserByID implements types.UserStore.
func (m *mockUserStore) GetUserByID(id int) (*types.User, error) {
	panic("unimplemented")
}
func (m *mockUserStore) GetReceiptByID(userId int, receiptId int) error {
	return nil
}
