package table

import (
	"context"

	"github.com/tiptop-co/backend/internal/model/table"
)

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
