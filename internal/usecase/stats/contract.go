package stats

import (
	"context"

	statsmodel "github.com/tiptop-co/backend/internal/model/stats"
)

type Usecase interface {
	VenueStats(ctx context.Context, venueID string) (*statsmodel.VenueStats, error)
	GlobalStats(ctx context.Context) (*statsmodel.GlobalStats, error)
}

type Repository interface {
	VenueStats(ctx context.Context, venueID string) (*statsmodel.VenueStats, error)
	GlobalStats(ctx context.Context) (*statsmodel.GlobalStats, error)
}
