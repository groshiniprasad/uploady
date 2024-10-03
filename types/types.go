package types

import (
	"net/http"
	"time"
)

type Config struct {
	Port       string
	DBUser     string
	DBPassword string
	DBAddress  string
	DBName     string
}

type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"createdAt"`
}

type UserStore interface {
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id int) (*User, error)
	CreateUser(User) error
}

type CreateProductPayload struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	Image       string  `json:"image"`
	Price       float64 `json:"price" validate:"required"`
	Quantity    int     `json:"quantity" validate:"required"`
}

type RegisterUserPayload struct {
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=4,max=13"`
}

type LoginUserPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type Receipt struct {
	ID          int       `json:"id"`
	UserID      int       `json:"userID"`
	Name        string    `json:"name"`
	Amount      float64   `json:"amount"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	ImagePath   string    `json:"imagePath"`
	CreatedAt   time.Time `json:"createdAt"`
}

type ReceiptStore interface {
	GetReceiptByName(name string, userId int) (*User, error)
	CreateReceipt(Receipt) (int, error)
	GetReceiptByID(receiptId int, userId int) (*Receipt, error)
}

type CreateReceiptPayload struct {
	ID          int       `json:"id"`
	UserID      int       `json:"userID"`
	Name        string    `json:"name" validate:"required"`
	Amount      float64   `json:"amount" validate:"required,gt=0"`
	Date        time.Time `json:"date" validate:"required"`
	Description string    `json:"description"`
}

// Task represents a heavy task for the worker pool
type Task struct {
	Request  *http.Request
	Response http.ResponseWriter
	JobType  string // Defines what kind of heavy task (e.g., image processing)
	Payload  interface{}
}

type WorkerPool struct {
	taskQueue chan Task
	quit      chan bool
}

type ResizeTaskPayload struct {
	ImagePath string
	Width     int
	Height    int
	Response  http.ResponseWriter
}

type CropTaskPayload struct {
	ImagePath string
	Width     int
	Height    int
	Response  http.ResponseWriter
	X         int
	Y         int
}
