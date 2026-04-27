package table

import (
	"context"
	"errors"

	"github.com/tiptop-co/backend/internal/model/table"
)

var (
	ErrTableIsOccupied     = errors.New("table is occupied")
	ErrInvalidTableSession = errors.New("invalid table session")
)

type TableUsecase interface {
	Create(ctx context.Context, venueID string, tableNumber int) (_ *table.Table, err error)

	GetByID(ctx context.Context, tableID string) (_ *table.Table, err error)
	GetByFilters(ctx context.Context, filters *table.TableFilters) (_ []*table.Table, err error)

	AssignWaiter(ctx context.Context, tableID string, waiterID string) (err error)
	UnassignWaiter(ctx context.Context, tableID string) (err error)

	UpdateStatus(ctx context.Context, tableID string, status table.Status) (err error)
	Update(ctx context.Context, table *table.Table) (err error)

	Delete(ctx context.Context, tableID string) (err error)

	ValidateSessionToken(ctx context.Context, sessionToken string) (_ *table.Table, err error)
}

type TableRepository interface {
	Create(ctx context.Context, table *table.Table) error

	GetByID(ctx context.Context, tableID string) (*table.Table, error)
	GetByFilters(ctx context.Context, filters *table.TableFilters) ([]*table.Table, error)
	Update(ctx context.Context, table *table.Table) error

	Delete(ctx context.Context, id string) error
}

type TokenGenerator interface {
	GenerateToken(size int) (string, error)
}
