package receipts

import (
	"sync"
)

type Receipt struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
}

var (
	receipts = make(map[string]Receipt)
	mutex    = &sync.Mutex{}
)

func SaveReceipt(receipt Receipt) {
	mutex.Lock()
	defer mutex.Unlock()
	receipts[receipt.ID] = receipt
}

func GetAllReceipts() []Receipt {
	mutex.Lock()
	defer mutex.Unlock()
	var result []Receipt
	for _, receipt := range receipts {
		result = append(result, receipt)
	}
	return result
}

func DeleteReceipt(id string) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(receipts, id)
}
