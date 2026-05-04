package waiter_request

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/tiptop-co/backend/internal/model"
	wrmodel "github.com/tiptop-co/backend/internal/model/waiter_request"
	"github.com/tiptop-co/backend/pkg/errwrap"
)

const errPrefix = "[WAITER_REQUEST USECASE]"

var activeStatuses = []wrmodel.Status{wrmodel.StatusPending, wrmodel.StatusAccepted}

type Service struct {
	repo     Repository
	tables   TableLookup
	assigner WaiterAssigner
}

func NewService(repo Repository, tables TableLookup, assigner WaiterAssigner) *Service {
	return &Service{repo: repo, tables: tables, assigner: assigner}
}

func (s *Service) CanCall(ctx context.Context, tableID string) (_ bool, err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" can call", err) }()

	if tableID == "" {
		return false, model.ErrValidation
	}
	active, err := s.repo.GetByFilters(ctx, &wrmodel.Filters{
		TableID:  &tableID,
		Statuses: activeStatuses,
	})
	if err != nil {
		return false, err
	}
	return len(active) == 0, nil
}

func (s *Service) GetByTable(ctx context.Context, tableID string) (_ []*wrmodel.Request, err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" get by table", err) }()

	if tableID == "" {
		return nil, model.ErrValidation
	}
	return s.repo.GetByFilters(ctx, &wrmodel.Filters{TableID: &tableID})
}

func (s *Service) GetByWaiter(ctx context.Context, waiterID string) (_ []*wrmodel.Request, err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" get by waiter", err) }()

	if waiterID == "" {
		return nil, model.ErrValidation
	}
	return s.repo.GetByFilters(ctx, &wrmodel.Filters{
		WaiterID: &waiterID,
		Statuses: activeStatuses,
	})
}

func (s *Service) Create(ctx context.Context, tableID string) (_ *wrmodel.Request, err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" create", err) }()

	if tableID == "" {
		return nil, model.ErrValidation
	}

	can, err := s.CanCall(ctx, tableID)
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, ErrCallNotAllowed
	}

	t, err := s.tables.GetByID(ctx, tableID)
	if err != nil {
		return nil, err
	}

	var waiterID *string
	if t.WaiterID != nil {
		waiterID = t.WaiterID
	} else {
		assigned, err := s.assigner.LeastLoadedWaiter(ctx, t.VenueID)
		if err != nil {
			return nil, err
		}
		waiterID = assigned
	}

	venueID := t.VenueID
	req := &wrmodel.Request{
		ID:          uuid.New().String(),
		TableID:     tableID,
		TableNumber: t.Number,
		VenueID:     &venueID,
		WaiterID:    waiterID,
		Status:      wrmodel.StatusPending,
		CreatedAt:   time.Now().UTC(),
	}
	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}
	return req, nil
}

func (s *Service) Accept(ctx context.Context, requestID, waiterID string) (_ *wrmodel.Request, err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" accept", err) }()

	if requestID == "" || waiterID == "" {
		return nil, model.ErrValidation
	}

	req, err := s.repo.GetByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req.Status != wrmodel.StatusPending {
		return nil, model.ErrValidation
	}
	if req.WaiterID != nil && *req.WaiterID != waiterID {
		return nil, ErrNotOwnRequest
	}

	if req.WaiterID == nil {
		if err := s.repo.UpdateWaiter(ctx, requestID, waiterID); err != nil {
			return nil, err
		}
		req.WaiterID = &waiterID
	}
	if err := s.repo.UpdateStatus(ctx, requestID, wrmodel.StatusAccepted); err != nil {
		return nil, err
	}
	req.Status = wrmodel.StatusAccepted
	return req, nil
}
