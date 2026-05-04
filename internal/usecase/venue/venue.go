package venue

import (
	"context"

	"github.com/tiptop-co/backend/internal/model"
	"github.com/tiptop-co/backend/internal/model/venue"
	"github.com/tiptop-co/backend/pkg/errwrap"
)

const errPrefix = "[VENUE USECASE]"

type VenueService struct {
	repo VenueRepository
}

func NewVenueService(repo VenueRepository) *VenueService {
	return &VenueService{repo: repo}
}

func (s *VenueService) GetAll(ctx context.Context) (_ []*venue.Venue, err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" get all", err) }()
	return s.repo.GetAll(ctx)
}

func (s *VenueService) AssignManager(ctx context.Context, venueID, managerID string) (err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" assign manager", err) }()
	if venueID == "" || managerID == "" {
		return model.ErrValidation
	}
	return s.repo.AssignManager(ctx, venueID, managerID)
}

func (s *VenueService) GetByManager(ctx context.Context, managerID string) (_ *venue.Venue, err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" get by manager", err) }()

	if managerID == "" {
		return nil, model.ErrValidation
	}
	return s.repo.GetByManager(ctx, managerID)
}

func (s *VenueService) UpdateByManager(ctx context.Context, managerID string, in *venue.UpdateInput) (_ *venue.Venue, err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" update by manager", err) }()

	if managerID == "" || in == nil || in.Name == "" || in.Address == "" {
		return nil, model.ErrValidation
	}

	v, err := s.repo.GetByManager(ctx, managerID)
	if err != nil {
		return nil, err
	}

	v.Name = in.Name
	v.Address = in.Address
	v.Description = in.Description
	v.BankAccount = in.BankAccount

	if err := s.repo.Update(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}
