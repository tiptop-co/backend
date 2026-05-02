package user

import (
	"time"

	"github.com/tiptop-co/backend/internal/model/authz"
)

type User struct {
	ID        string         `json:"id"         db:"id"`
	FirstName string         `json:"first_name" db:"first_name"`
	LastName  string         `json:"last_name"  db:"last_name"`
	Phone     string         `json:"phone"      db:"phone"`
	Role      authz.UserRole `json:"role"       db:"role"`
	VenueID   *string        `json:"venue_id,omitempty"   db:"venue_id"`
	VenueName *string        `json:"venue_name,omitempty" db:"venue_name"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
}

type Filters struct {
	ID      *string
	Phone   *string
	Role    *authz.UserRole
	VenueID *string
}
