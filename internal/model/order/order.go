package order

import "time"

type Status string

const (
	StatusUnspecified Status = "unspecified"
	StatusActive      Status = "active"
	StatusCompleted   Status = "completed"
)

type ItemStatus string

const (
	ItemStatusUnspecified ItemStatus = "unspecified"
	ItemStatusUnpaid      ItemStatus = "unpaid"
	ItemStatusPaid        ItemStatus = "paid"
)

type CompletedSummary struct {
	OrderID     string    `json:"order_id"`
	TableID     string    `json:"table_id"`
	TableNumber int       `json:"table_number"`
	TotalAmount int       `json:"total_amount"`
	ItemsCount  int       `json:"items_count"`
	CompletedAt time.Time `json:"completed_at"`
}

type Item struct {
	ID         string     `json:"id"          db:"id"`
	OrderID    string     `json:"order_id"    db:"order_id"`
	DishID     string     `json:"dish_id"     db:"dish_id"`
	DishName   string     `json:"dish_name"   db:"dish_name"`
	Quantity   int        `json:"quantity"    db:"quantity"`
	Price      int        `json:"price"       db:"price"`
	Status     ItemStatus `json:"status"      db:"status"`
	AddedLater bool       `json:"added_later" db:"added_later"`
}

type Order struct {
	ID          string    `json:"id"           db:"id"`
	TableID     string    `json:"table_id"     db:"table_id"`
	WaiterID    *string   `json:"waiter_id"    db:"waiter_id"`
	Status      Status    `json:"status"       db:"status"`
	Items       []Item    `json:"items"`
	TotalAmount int       `json:"total_amount" db:"total_amount"`
	PaidAmount  int       `json:"paid_amount"  db:"paid_amount"`
	Wishes      *string   `json:"wishes"       db:"wishes"`
	CreatedAt   time.Time `json:"created_at"   db:"created_at"`
}

type CreateItemInput struct {
	DishID   string
	Quantity int
}
