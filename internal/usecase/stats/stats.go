package stats

import (
	"context"

	"github.com/tiptop-co/backend/internal/model"
	statsmodel "github.com/tiptop-co/backend/internal/model/stats"
	"github.com/tiptop-co/backend/pkg/errwrap"
)

const errPrefix = "[STATS USECASE]"

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) VenueStats(ctx context.Context, venueID string) (_ *statsmodel.VenueStats, err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" venue stats", err) }()

	if venueID == "" {
		return nil, model.ErrValidation
	}
	return s.repo.VenueStats(ctx, venueID)
}

func (s *Service) GlobalStats(ctx context.Context) (_ *statsmodel.GlobalStats, err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" global stats", err) }()
	return s.repo.GlobalStats(ctx)
}
