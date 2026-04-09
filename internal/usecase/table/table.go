package table

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/tiptop-co/backend/internal/model"
	"github.com/tiptop-co/backend/internal/model/table"
	"github.com/tiptop-co/backend/pkg/errwrap"
)

var (
	ErrTableIsOpen           = errors.New("table is open")
	ErrTableSessionIsExpired = errors.New("table session is expired")
)

type TableUsecase struct {
	sizeQRToken      int
	sizeSessionToken int

	repo TableRepository
	gen  TokenGenerator
}

func NewTableUsecase(sizeQRToken int, sizeSessionToken int,
	repo TableRepository, gen TokenGenerator) *TableUsecase {
	return &TableUsecase{
		sizeQRToken:      sizeQRToken,
		sizeSessionToken: sizeSessionToken,

		repo: repo,
		gen:  gen,
	}
}

func (u *TableUsecase) Create(ctx context.Context, venueID string, tableNumber int) (_ *table.Table, err error) {
	defer func() {
		err = errwrap.WrapWithMsg("[TABLE USECASE] create", err)
	}()

	if venueID == "" || tableNumber < 0 {
		return nil, model.ErrInvalidParams
	}

	qrToken, err := u.gen.GenerateToken(u.sizeQRToken)
	if err != nil {
		return nil, fmt.Errorf("generate qr token: %w", err)
	}

	sessionToken, err := u.gen.GenerateToken(u.sizeSessionToken)
	if err != nil {
		return nil, fmt.Errorf("generate session token: %w", err)
	}

	table := &table.Table{
		ID:           uuid.New().String(),
		VenueID:      venueID,
		Number:       tableNumber,
		QRToken:      qrToken,
		SessionToken: sessionToken,
		Status:       table.StatusClosed,
	}

	if err = u.repo.Create(ctx, table); err != nil {
		return nil, fmt.Errorf("failed to create: %w", err)
	}

	return table, nil
}

func (u *TableUsecase) GetByID(ctx context.Context, tableID string) (_ *table.Table, err error) {
	defer func() {
		err = errwrap.WrapWithMsg("[TABLE USECASE] get by id", err)
	}()

	return u.repo.GetByID(ctx, tableID)
}

func (u *TableUsecase) GetByFilters(ctx context.Context, filters *table.TableFilters) (_ []*table.Table, err error) {
	defer func() {
		err = errwrap.WrapWithMsg("[TABLE USECASE] get by filters", err)
	}()

	return u.repo.GetByFilters(ctx, filters)
}

func (u *TableUsecase) AssignWaiter(ctx context.Context, tableID string, waiterID string) (err error) {
	defer func() {
		err = errwrap.WrapWithMsg("[TABLE USECASE] assignWaiter", err)
	}()

	return u.updateWaiter(ctx, tableID, &waiterID)
}

func (u *TableUsecase) UnassignWaiter(ctx context.Context, tableID string) (err error) {
	defer func() {
		err = errwrap.WrapWithMsg("[TABLE USECASE] unassignWaiter", err)
	}()

	return u.updateWaiter(ctx, tableID, nil)
}

func (u *TableUsecase) updateWaiter(ctx context.Context, tableID string, waiterID *string) error {
	t, err := u.repo.GetByID(ctx, tableID)
	if err != nil {
		return fmt.Errorf("failed to get table: %w", err)
	}
	t.WaiterID = waiterID

	if err = u.repo.Update(ctx, t); err != nil {
		return fmt.Errorf("failed to update table's waiter: %w", err)
	}

	return nil
}

func (u *TableUsecase) Close(ctx context.Context, tableID string) (err error) {
	defer func() {
		err = errwrap.WrapWithMsg("[TABLE USECASE] close", err)
	}()

	t, err := u.repo.GetByID(ctx, tableID)
	if err != nil {
		return fmt.Errorf("failed to get table: %w", err)
	}

	newSessionToken, err := u.gen.GenerateToken(u.sizeSessionToken)
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	t.Status = table.StatusClosed
	t.SessionToken = newSessionToken
	t.WaiterID = nil

	if err = u.repo.Update(ctx, t); err != nil {
		return fmt.Errorf("failed to update table info: %w", err)
	}

	return nil
}

func (u *TableUsecase) Delete(ctx context.Context, tableID string) (err error) {
	defer func() {
		err = errwrap.WrapWithMsg("[TABLE USECASE] delete", err)
	}()

	t, err := u.repo.GetByID(ctx, tableID)
	if err != nil {
		return fmt.Errorf("failed to get table: %w", err)
	}

	if t.Status == table.StatusOpen {
		return ErrTableIsOpen
	}

	if err = u.repo.Delete(ctx, tableID); err != nil {
		return fmt.Errorf("failed: %w", err)
	}

	return nil
}

func (u *TableUsecase) ValidateSessionToken(ctx context.Context, tableID string, sessionToken string) (err error) {
	defer func() {
		err = errwrap.WrapWithMsg("[TABLE USECASE] validateSessionToken", err)
	}()

	t, err := u.repo.GetByID(ctx, tableID)
	if err != nil {
		return fmt.Errorf("failed to get table: %w", err)
	}

	if t.SessionToken != sessionToken {
		return ErrTableSessionIsExpired
	}

	return nil
}
