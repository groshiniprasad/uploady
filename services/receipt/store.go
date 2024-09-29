package receipt

import (
	"database/sql"
	"fmt"

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

	// Get the last inserted ID
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve last insert ID: %w", err)
	}

	return int(id), nil
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
