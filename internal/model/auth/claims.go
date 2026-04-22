package auth

import "github.com/tiptop-co/backend/internal/model/authz"

type Claims struct {
	UserID   string         `json:"user_id" db:"id"`
	UserRole authz.UserRole `json:"user_role" db:"role"`
	VenueID  string         `json:"venue_id" db:"venue_id"`
}
