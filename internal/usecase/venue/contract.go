package venue

import (
	"context"

	"github.com/tiptop-co/backend/internal/model/venue"
)

type VenueUsecase interface {
	GetAll(ctx context.Context) ([]*venue.Venue, error)
	GetByManager(ctx context.Context, managerID string) (*venue.Venue, error)
	UpdateByManager(ctx context.Context, managerID string, in *venue.UpdateInput) (*venue.Venue, error)
	AssignManager(ctx context.Context, venueID, managerID string) error
}

type VenueRepository interface {
	GetAll(ctx context.Context) ([]*venue.Venue, error)
	GetByID(ctx context.Context, id string) (*venue.Venue, error)
	GetByManager(ctx context.Context, managerID string) (*venue.Venue, error)
	Update(ctx context.Context, v *venue.Venue) error
	AssignManager(ctx context.Context, venueID, managerID string) error
}
