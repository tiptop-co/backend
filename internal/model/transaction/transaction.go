package transaction

import "time"

type Status string

const (
	StatusUnspecified Status = "unspecified"
	StatusPending     Status = "pending"
	StatusSuccess     Status = "success"
	StatusFailed      Status = "failed"
)

type Transaction struct {
	ID         string    `json:"id"          db:"id"`
	OrderID    string    `json:"order_id"    db:"order_id"`
	Amount     int       `json:"amount"      db:"amount"`
	TipsAmount int       `json:"tips_amount" db:"tips_amount"`
	Status     Status    `json:"status"      db:"status"`
	ItemIDs    []string  `json:"item_ids"`
	CreatedAt  time.Time `json:"created_at"  db:"created_at"`
}
