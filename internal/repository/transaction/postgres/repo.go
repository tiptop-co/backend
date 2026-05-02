package transaction_postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tiptop-co/backend/internal/model/transaction"
)

type TransactionRepository struct {
	db *pgxpool.Pool
}

func NewTransactionRepository(db *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Pool() *pgxpool.Pool { return r.db }

func (r *TransactionRepository) Create(ctx context.Context, tx pgx.Tx, t *transaction.Transaction) error {
	const query = `
		INSERT INTO transactions (id, order_id, amount, tips_amount, status, created_at)
		VALUES (@id, @order_id, @amount, @tips_amount, @status, @created_at)
	`
	args := pgx.NamedArgs{
		"id":          t.ID,
		"order_id":    t.OrderID,
		"amount":      t.Amount,
		"tips_amount": t.TipsAmount,
		"status":      string(t.Status),
		"created_at":  t.CreatedAt,
	}
	if _, err := tx.Exec(ctx, query, args); err != nil {
		return fmt.Errorf("transaction repository create: %w", err)
	}
	return nil
}

func (r *TransactionRepository) LinkItems(ctx context.Context, tx pgx.Tx, transactionID string, itemIDs []string) error {
	const query = `INSERT INTO transaction_items (transaction_id, order_item_id) VALUES ($1, $2)`
	for _, id := range itemIDs {
		if _, err := tx.Exec(ctx, query, transactionID, id); err != nil {
			return fmt.Errorf("transaction repository link item: %w", err)
		}
	}
	return nil
}

func (r *TransactionRepository) UpdateStatus(ctx context.Context, tx pgx.Tx, id string, status transaction.Status) error {
	const query = `UPDATE transactions SET status = $1 WHERE id = $2`
	if _, err := tx.Exec(ctx, query, string(status), id); err != nil {
		return fmt.Errorf("transaction repository update status: %w", err)
	}
	return nil
}

func (r *TransactionRepository) SumTipsByWaiterToday(ctx context.Context, waiterID string) (int, error) {
	now := time.Now().UTC()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.Add(24 * time.Hour)

	const query = `
		SELECT COALESCE(SUM(t.tips_amount), 0)
		FROM transactions t
		JOIN orders o ON o.id = t.order_id
		WHERE t.status = 'success'
		  AND o.waiter_id = $1
		  AND t.created_at >= $2 AND t.created_at < $3
	`
	var total int
	err := r.db.QueryRow(ctx, query, waiterID, dayStart, dayEnd).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("transaction repository sum tips today: %w", err)
	}
	return total, nil
}
