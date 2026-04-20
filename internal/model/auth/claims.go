package auth

type Claims struct {
	UserID   string   `json:"user_id" db:"id"`
	UserRole UserRole `json:"user_role" db:"role"`
	VenueID  string   `json:"venue_id" db:"venue_id"`
}
