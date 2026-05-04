package table_close

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tiptop-co/backend/internal/model/order"
	"github.com/tiptop-co/backend/internal/model/table"
)

var ErrUnpaidItems = errors.New("table has unpaid items")

type Usecase interface {
	Close(ctx context.Context, tableID string) (*table.Table, error)
}

type TableRepo interface {
	GetByID(ctx context.Context, id string) (*table.Table, error)
	Update(ctx context.Context, t *table.Table) error
}

type OrderRepo interface {
	Pool() *pgxpool.Pool
	GetActiveByTable(ctx context.Context, tableID string) (*order.Order, error)
	MarkCompleted(ctx context.Context, tx pgx.Tx, orderID string) error
}

type TokenGenerator interface {
	GenerateToken(size int) (string, error)
}
