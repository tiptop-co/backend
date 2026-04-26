package table

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tiptop-co/backend/internal/config"
	"github.com/tiptop-co/backend/internal/model"
	"github.com/tiptop-co/backend/internal/model/table"
	"github.com/tiptop-co/backend/pkg/errwrap"
)

type TableService struct {
	sizeQRToken      int
	sizeSessionToken int

	repo TableRepository
	gen  TokenGenerator
}

func NewTableService(cfg *config.TableServiceConfig,
	repo TableRepository, gen TokenGenerator) *TableService {
	return &TableService{
		sizeQRToken:      cfg.QRTokenSize,
		sizeSessionToken: cfg.SessionTokenSize,

		repo: repo,
		gen:  gen,
	}
}

func (s *TableService) Create(ctx context.Context, venueID string, tableNumber int) (_ *table.Table, err error) {
	defer func() {
		err = errwrap.WrapMsg("[TABLE USECASE] create", err)
	}()

	if venueID == "" || tableNumber < 0 {
		return nil, model.ErrValidation
	}

	qrToken, err := s.gen.GenerateToken(s.sizeQRToken)
	if err != nil {
		return nil, fmt.Errorf("generate qr token: %w", err)
	}

	sessionToken, err := s.gen.GenerateToken(s.sizeSessionToken)
	if err != nil {
		return nil, fmt.Errorf("generate session token: %w", err)
	}

	table := &table.Table{
		ID:           uuid.New().String(),
		VenueID:      venueID,
		Number:       tableNumber,
		QRToken:      qrToken,
		SessionToken: sessionToken,
		Status:       table.StatusFree,
	}

	if err = s.repo.Create(ctx, table); err != nil {
		return nil, fmt.Errorf("failed to create: %w", err)
	}

	return table, nil
}

func (s *TableService) GetByID(ctx context.Context, tableID string) (_ *table.Table, err error) {
	defer func() {
		err = errwrap.WrapMsg("[TABLE USECASE] get by id", err)
	}()

	return s.repo.GetByID(ctx, tableID)
}

func (s *TableService) GetByFilters(ctx context.Context, filters *table.TableFilters) (_ []*table.Table, err error) {
	defer func() {
		err = errwrap.WrapMsg("[TABLE USECASE] get by filters", err)
	}()

	return s.repo.GetByFilters(ctx, filters)
}

func (s *TableService) AssignWaiter(ctx context.Context, tableID string, waiterID string) (err error) {
	defer func() {
		err = errwrap.WrapMsg("[TABLE USECASE] assignWaiter", err)
	}()

	return s.updateWaiter(ctx, tableID, &waiterID)
}

func (s *TableService) UnassignWaiter(ctx context.Context, tableID string) (err error) {
	defer func() {
		err = errwrap.WrapMsg("[TABLE USECASE] unassignWaiter", err)
	}()

	return s.updateWaiter(ctx, tableID, nil)
}

func (s *TableService) updateWaiter(ctx context.Context, tableID string, waiterID *string) error {
	t, err := s.repo.GetByID(ctx, tableID)
	if err != nil {
		return fmt.Errorf("failed to get table: %w", err)
	}
	t.WaiterID = waiterID

	if err = s.repo.Update(ctx, t); err != nil {
		return fmt.Errorf("failed to update table's waiter: %w", err)
	}

	return nil
}

func (s *TableService) UpdateStatus(ctx context.Context, tableID string, status table.Status) (err error) {
	defer func() {
		err = errwrap.WrapMsg("[TABLE USECASE] update status", err)
	}()

	t, err := s.repo.GetByID(ctx, tableID)
	if err != nil {
		return fmt.Errorf("failed to get table: %w", err)
	}

	t.Status = status
	if status == table.StatusFree {
		newSessionToken, err := s.gen.GenerateToken(s.sizeSessionToken)
		if err != nil {
			return fmt.Errorf("failed to generate token: %w", err)
		}

		t.SessionToken = newSessionToken
		t.WaiterID = nil
	}

	if err = s.repo.Update(ctx, t); err != nil {
		return fmt.Errorf("failed to update table info: %w", err)
	}

	return nil
}

func (s *TableService) Delete(ctx context.Context, tableID string) (err error) {
	defer func() {
		err = errwrap.WrapMsg("[TABLE USECASE] delete", err)
	}()

	t, err := s.repo.GetByID(ctx, tableID)
	if err != nil {
		return fmt.Errorf("failed to get table: %w", err)
	}

	if t.Status == table.StatusPaid || t.Status == table.StatusUnpaid {
		return ErrTableIsOccupied
	}

	if err = s.repo.Delete(ctx, tableID); err != nil {
		return fmt.Errorf("failed: %w", err)
	}

	return nil
}

func (s *TableService) ValidateSessionToken(ctx context.Context, sessionToken string) (_ *table.Table, err error) {
	defer func() {
		err = errwrap.WrapMsg("[TABLE USECASE] validate session token", err)
	}()

	if sessionToken == "" {
		return nil, model.ErrValidation
	}

	tables, err := s.repo.GetByFilters(ctx, &table.TableFilters{
		Session: &sessionToken,
	})
	if err != nil {
		return nil, errwrap.WrapMsg("failed to get table", err)
	}
	if len(tables) == 0 {
		return nil, ErrInvalidTableSession
	}

	t := tables[0]
	if t.SessionToken != sessionToken {
		return nil, ErrInvalidTableSession
	}

	return t, nil
}

func (s *TableService) Update(ctx context.Context, table *table.Table) (err error) {
	defer func() {
		err = errwrap.WrapMsg("[TABLE USECASE] update", err)
	}()

	return s.repo.Update(ctx, table)
}
