package table

type Status string

const (
	StatusOpen   Status = "open"
	StatusClosed Status = "closed"
)

type Table struct {
	ID      string `json:"id" db:"id"`
	Number  int    `json:"number" db:"number"` // the table number in a particular venue. This is necessary for the convenience of employees.
	Status  Status `json:"status" db:"status"`
	VenueID string `json:"venue_id" db:"venue_id"` // ID of cafe, restaurant, etc.
	QRToken string `json:"qr_token" db:"qr_token"`

	// The session is necessary in order to avoid the "ghost problem".
	// Previous visitors, whose tabs are still open, will be able to interfere
	// with the order of current visitors (for example, to order something).
	//
	// After the table is closed, the session_token will be updated,
	// so that the table cart can only be opened via a QR code.
	SessionToken string `json:"session_token" db:"session_token"`

	WaiterID *string `json:"waiter_id" db:"waiter_id"`
}

type TableFilters struct {
	TableID  *string
	QRToken  *string
	WaiterID *string
	VenueID  *string
}
