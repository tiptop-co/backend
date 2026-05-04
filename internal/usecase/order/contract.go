package order

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tiptop-co/backend/internal/model/dish"
	"github.com/tiptop-co/backend/internal/model/order"
	"github.com/tiptop-co/backend/internal/model/table"
)

var (
	ErrEmptyItems     = errors.New("order must have items")
	ErrDishWrongVenue = errors.New("dish does not belong to table venue")
	ErrTableNotFound  = errors.New("table not found")
)

type OrderUsecase interface {
	GetByTable(ctx context.Context, tableID string) (*order.Order, error)
	Create(ctx context.Context, tableID string, items []order.CreateItemInput, wishes *string) (*order.Order, error)
	GetCompletedByWaiter(ctx context.Context, waiterID string, limit int) ([]*order.CompletedSummary, error)
}

type OrderRepository interface {
	Pool() *pgxpool.Pool
	Create(ctx context.Context, tx pgx.Tx, o *order.Order) error
	GetActiveByTable(ctx context.Context, tableID string) (*order.Order, error)
	GetByID(ctx context.Context, id string) (*order.Order, error)
	GetItemsByOrder(ctx context.Context, orderID string) ([]order.Item, error)
	AddItems(ctx context.Context, tx pgx.Tx, items []order.Item) error
	UpdateTotals(ctx context.Context, tx pgx.Tx, orderID string, total, paid int) error
	UpdateWaiter(ctx context.Context, tx pgx.Tx, orderID string, waiterID *string) error
	MarkItemsPaid(ctx context.Context, tx pgx.Tx, itemIDs []string) error
	GetItemsByIDs(ctx context.Context, orderID string, itemIDs []string) ([]order.Item, error)
	MarkCompleted(ctx context.Context, tx pgx.Tx, orderID string) error
	GetCompletedByWaiter(ctx context.Context, waiterID string, limit int) ([]*order.CompletedSummary, error)
}

type DishLookup interface {
	GetDishByID(ctx context.Context, id string) (*dish.Dish, error)
}

type TableRepo interface {
	GetByID(ctx context.Context, id string) (*table.Table, error)
	Update(ctx context.Context, t *table.Table) error
}

type WaiterAssigner interface {
	LeastLoadedWaiter(ctx context.Context, venueID string) (*string, error)
}
