package waiter_request

import "time"

type Status string

const (
	StatusUnspecified Status = "unspecified"
	StatusPending     Status = "pending"
	StatusAccepted    Status = "accepted"
	StatusDone        Status = "done"
)

type Request struct {
	ID          string    `json:"id"           db:"id"`
	TableID     string    `json:"table_id"     db:"table_id"`
	TableNumber int       `json:"table_number"`
	VenueID     *string   `json:"venue_id"     db:"venue_id"`
	WaiterID    *string   `json:"waiter_id"    db:"waiter_id"`
	Status      Status    `json:"status"       db:"status"`
	CreatedAt   time.Time `json:"created_at"   db:"created_at"`
}

type Filters struct {
	TableID  *string
	WaiterID *string
	VenueID  *string
	Statuses []Status
}
