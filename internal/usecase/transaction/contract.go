package transaction

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tiptop-co/backend/internal/model/order"
	"github.com/tiptop-co/backend/internal/model/table"
	"github.com/tiptop-co/backend/internal/model/transaction"
)

var (
	ErrItemAlreadyPaid    = errors.New("item already paid")
	ErrItemNotInOrder     = errors.New("item does not belong to order")
	ErrPaymentFailed      = errors.New("payment failed")
	ErrEmptyItemList      = errors.New("at least one item required")
)

type TransactionUsecase interface {
	Create(ctx context.Context, orderID string, itemIDs []string, tipsAmount int) (*transaction.Transaction, error)
}

type OrderRepository interface {
	Pool() *pgxpool.Pool
	GetByID(ctx context.Context, id string) (*order.Order, error)
	GetItemsByIDs(ctx context.Context, orderID string, itemIDs []string) ([]order.Item, error)
	MarkItemsPaid(ctx context.Context, tx pgx.Tx, itemIDs []string) error
	UpdateTotals(ctx context.Context, tx pgx.Tx, orderID string, total, paid int) error
}

type TransactionRepository interface {
	Create(ctx context.Context, tx pgx.Tx, t *transaction.Transaction) error
	LinkItems(ctx context.Context, tx pgx.Tx, transactionID string, itemIDs []string) error
	UpdateStatus(ctx context.Context, tx pgx.Tx, id string, status transaction.Status) error
}

type TableRepo interface {
	GetByID(ctx context.Context, id string) (*table.Table, error)
	Update(ctx context.Context, t *table.Table) error
}

type PaymentGateway interface {
	Charge(ctx context.Context, transactionID string, amount int) (transaction.Status, error)
}
