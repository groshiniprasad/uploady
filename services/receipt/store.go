package receipt

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/groshiniprasad/uploady/types"
)

type Store struct {
	db *sql.DB
}

// GetReceiptByName implements types.ReceiptStore.
func (s *Store) GetReceiptByName(name string, userId int) (*types.User, error) {
	panic("unimplemented")
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateReceipt(receipt types.Receipt) (int, error) {
	// Execute the SQL insert statement
	res, err := s.db.Exec("INSERT INTO receipts (userId, name, amount, imagePath, date, description) VALUES (?, ?, ?, ?, ?, ?)",
		receipt.UserID, receipt.Name, receipt.Amount, receipt.ImagePath, receipt.Date, receipt.Description)
	if err != nil {
		return 0, fmt.Errorf("failed to create receipt: %w", err)
	}

	// Check if the receipt was successfully created
	log.Println("Receipt created successfully")

	// Get the last inserted ID
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve last insert ID: %w", err)
	}

	log.Printf("Receipt ID: %d\n", id)

	return int(id), nil
}

func (s *Store) GetReceiptByID(receiptId int, userId int) (*types.Receipt, error) {
	// Query the receipts table instead of users
	query := "SELECT id, userId, name, amount, date, imagePath FROM receipts WHERE id = ? AND userId = ?"
	row := s.db.QueryRow(query, receiptId, userId)
	log.Println("Querying the  with query:", row)

	r := new(types.Receipt)

	err := row.Scan(&r.ID, &r.UserID, &r.Name, &r.Amount, &r.Date, &r.ImagePath)

	if err == sql.ErrNoRows {
		// Handle the case where no rows are returned
		fmt.Println("No receipt found for the given ID and UserID")
		return nil, fmt.Errorf("receipt not found")
	} else if err != nil {
		fmt.Println("Error scanning row:", err)
		return nil, err
	}

	return r, nil
}

func (s *Store) GetReceiptsByName(receipt types.Receipt) ([]types.Receipt, error) {
	rows, err := s.db.Query("SELECT * FROM receipts WHERE userId = ? AND name LIKE ?", receipt.UserID, receipt.Name+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Initialize a slice to hold the results
	var receipts []types.Receipt

	// Iterate over the result rows and collect the receipts
	for rows.Next() {
		var foundReceipt types.Receipt
		if err := rows.Scan(&foundReceipt.ID, &foundReceipt.UserID, &foundReceipt.Name); err != nil {
			return nil, err
		}
		receipts = append(receipts, foundReceipt)
	}

	// If no receipts are found, return an appropriate error
	if len(receipts) == 0 {
		return nil, fmt.Errorf("no receipts found with name: %s", receipt.Name)
	}

	return receipts, nil
}
