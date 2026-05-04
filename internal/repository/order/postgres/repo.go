package order_postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tiptop-co/backend/internal/model"
	"github.com/tiptop-co/backend/internal/model/order"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Pool() *pgxpool.Pool { return r.db }

func (r *OrderRepository) GetCompletedByWaiter(ctx context.Context, waiterID string, limit int) ([]*order.CompletedSummary, error) {
	const query = `
		SELECT o.id, o.table_id, t.number, o.total_amount,
		       COALESCE((SELECT COUNT(*) FROM order_items oi WHERE oi.order_id = o.id), 0) AS items_count,
		       o.created_at
		FROM orders o
		JOIN tables t ON t.id = o.table_id
		WHERE o.waiter_id = $1 AND o.status = 'completed'
		ORDER BY o.created_at DESC
		LIMIT $2
	`
	rows, err := r.db.Query(ctx, query, waiterID, limit)
	if err != nil {
		return nil, fmt.Errorf("order repository get completed by waiter: %w", err)
	}
	defer rows.Close()

	var res []*order.CompletedSummary
	for rows.Next() {
		var s order.CompletedSummary
		if err := rows.Scan(&s.OrderID, &s.TableID, &s.TableNumber, &s.TotalAmount, &s.ItemsCount, &s.CompletedAt); err != nil {
			return nil, fmt.Errorf("scan completed order: %w", err)
		}
		res = append(res, &s)
	}
	return res, nil
}

func (r *OrderRepository) Create(ctx context.Context, tx pgx.Tx, o *order.Order) error {
	const query = `
		INSERT INTO orders (id, table_id, waiter_id, status, total_amount, paid_amount, wishes, created_at)
		VALUES (@id, @table_id, @waiter_id, @status, @total_amount, @paid_amount, @wishes, @created_at)
	`
	args := pgx.NamedArgs{
		"id":           o.ID,
		"table_id":     o.TableID,
		"waiter_id":    o.WaiterID,
		"status":       string(o.Status),
		"total_amount": o.TotalAmount,
		"paid_amount":  o.PaidAmount,
		"wishes":       o.Wishes,
		"created_at":   o.CreatedAt,
	}
	if _, err := exec(ctx, r.db, tx, query, args); err != nil {
		return fmt.Errorf("order repository create: %w", err)
	}
	return nil
}

func (r *OrderRepository) GetActiveByTable(ctx context.Context, tableID string) (*order.Order, error) {
	const query = `
		SELECT id, table_id, waiter_id, status, total_amount, paid_amount, wishes, created_at
		FROM orders
		WHERE table_id = $1 AND status = 'active'
		LIMIT 1
	`
	var o order.Order
	var statusStr string
	err := r.db.QueryRow(ctx, query, tableID).Scan(
		&o.ID, &o.TableID, &o.WaiterID, &statusStr, &o.TotalAmount, &o.PaidAmount, &o.Wishes, &o.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("order repository get active: %w", err)
	}
	o.Status = order.Status(statusStr)
	items, err := r.GetItemsByOrder(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items
	return &o, nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*order.Order, error) {
	const query = `
		SELECT id, table_id, waiter_id, status, total_amount, paid_amount, wishes, created_at
		FROM orders WHERE id = $1
	`
	var o order.Order
	var statusStr string
	err := r.db.QueryRow(ctx, query, id).Scan(
		&o.ID, &o.TableID, &o.WaiterID, &statusStr, &o.TotalAmount, &o.PaidAmount, &o.Wishes, &o.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("order repository get by id: %w", err)
	}
	o.Status = order.Status(statusStr)
	items, err := r.GetItemsByOrder(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items
	return &o, nil
}

func (r *OrderRepository) GetItemsByOrder(ctx context.Context, orderID string) ([]order.Item, error) {
	const query = `
		SELECT id, order_id, dish_id, dish_name, quantity, price, status, added_later
		FROM order_items WHERE order_id = $1
		ORDER BY added_later, id
	`
	rows, err := r.db.Query(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("order repository get items: %w", err)
	}
	defer rows.Close()

	var result []order.Item
	for rows.Next() {
		var (
			it        order.Item
			statusStr string
		)
		if err := rows.Scan(&it.ID, &it.OrderID, &it.DishID, &it.DishName, &it.Quantity, &it.Price, &statusStr, &it.AddedLater); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		it.Status = order.ItemStatus(statusStr)
		result = append(result, it)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("rows: %w", rows.Err())
	}
	return result, nil
}

func (r *OrderRepository) AddItems(ctx context.Context, tx pgx.Tx, items []order.Item) error {
	const query = `
		INSERT INTO order_items (id, order_id, dish_id, dish_name, quantity, price, status, added_later)
		VALUES (@id, @order_id, @dish_id, @dish_name, @quantity, @price, @status, @added_later)
	`
	for _, it := range items {
		args := pgx.NamedArgs{
			"id":          it.ID,
			"order_id":    it.OrderID,
			"dish_id":     it.DishID,
			"dish_name":   it.DishName,
			"quantity":    it.Quantity,
			"price":       it.Price,
			"status":      string(it.Status),
			"added_later": it.AddedLater,
		}
		if _, err := exec(ctx, r.db, tx, query, args); err != nil {
			return fmt.Errorf("order repository add item: %w", err)
		}
	}
	return nil
}

func (r *OrderRepository) UpdateTotals(ctx context.Context, tx pgx.Tx, orderID string, total, paid int) error {
	const query = `UPDATE orders SET total_amount = $1, paid_amount = $2 WHERE id = $3`
	if _, err := exec(ctx, r.db, tx, query, total, paid, orderID); err != nil {
		return fmt.Errorf("order repository update totals: %w", err)
	}
	return nil
}

func (r *OrderRepository) MarkItemsPaid(ctx context.Context, tx pgx.Tx, itemIDs []string) error {
	if len(itemIDs) == 0 {
		return nil
	}
	const query = `UPDATE order_items SET status = 'paid' WHERE id = ANY($1) AND status = 'unpaid'`
	if _, err := exec(ctx, r.db, tx, query, itemIDs); err != nil {
		return fmt.Errorf("order repository mark items paid: %w", err)
	}
	return nil
}

func (r *OrderRepository) GetItemsByIDs(ctx context.Context, orderID string, itemIDs []string) ([]order.Item, error) {
	if len(itemIDs) == 0 {
		return nil, nil
	}
	const query = `
		SELECT id, order_id, dish_id, dish_name, quantity, price, status, added_later
		FROM order_items WHERE order_id = $1 AND id = ANY($2)
	`
	rows, err := r.db.Query(ctx, query, orderID, itemIDs)
	if err != nil {
		return nil, fmt.Errorf("order repository get items by ids: %w", err)
	}
	defer rows.Close()

	var result []order.Item
	for rows.Next() {
		var (
			it        order.Item
			statusStr string
		)
		if err := rows.Scan(&it.ID, &it.OrderID, &it.DishID, &it.DishName, &it.Quantity, &it.Price, &statusStr, &it.AddedLater); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		it.Status = order.ItemStatus(statusStr)
		result = append(result, it)
	}
	return result, nil
}

func (r *OrderRepository) MarkCompleted(ctx context.Context, tx pgx.Tx, orderID string) error {
	const query = `UPDATE orders SET status = 'completed' WHERE id = $1`
	if _, err := exec(ctx, r.db, tx, query, orderID); err != nil {
		return fmt.Errorf("order repository mark completed: %w", err)
	}
	return nil
}

func (r *OrderRepository) UpdateWaiter(ctx context.Context, tx pgx.Tx, orderID string, waiterID *string) error {
	const query = `UPDATE orders SET waiter_id = $1 WHERE id = $2`
	if _, err := exec(ctx, r.db, tx, query, waiterID, orderID); err != nil {
		return fmt.Errorf("order repository update waiter: %w", err)
	}
	return nil
}

func exec(ctx context.Context, pool *pgxpool.Pool, tx pgx.Tx, sql string, args ...any) (any, error) {
	if tx != nil {
		return tx.Exec(ctx, sql, args...)
	}
	return pool.Exec(ctx, sql, args...)
}
