package table_postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tiptop-co/backend/internal/model"
	"github.com/tiptop-co/backend/internal/model/table"
)

var (
	ErrBuildQuery = errors.New("failed to build query")
	ErrNoFilters  = errors.New("filters must be in the request")
)

type TableRepository struct {
	db *pgxpool.Pool
}

func NewTableRepository(db *pgxpool.Pool) *TableRepository {
	return &TableRepository{db: db}
}

func (r *TableRepository) Create(ctx context.Context, t *table.Table) error {
	query := `
		INSERT INTO tables (id, venue_id, number, qr_token, session_token, status, waiter_id, order_id)
		VALUES (@id, @venue_id, @number, @qr_token, @session_token, @status, @waiter_id, @order_id)
	`

	args := pgx.NamedArgs{
		"id":            t.ID,
		"venue_id":      t.VenueID,
		"number":        t.Number,
		"qr_token":      t.QRToken,
		"session_token": t.SessionToken,
		"status":        t.Status,
		"waiter_id":     t.WaiterID,
		"order_id":      t.OrderID,
	}

	_, err := r.db.Exec(ctx, query, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			err = fmt.Errorf("%w: %w", model.ErrAlreadyExists, err)
		}
		return fmt.Errorf("table repository create: %w", err)
	}

	return nil
}

func (r *TableRepository) Update(ctx context.Context, t *table.Table) error {
	query := `
		UPDATE tables
		SET
			number        = @number,
			qr_token      = @qr_token,
			session_token = @session_token,
			status        = @status,
			waiter_id     = @waiter_id,
			order_id 	  = @order_id
		WHERE id = @id
	`

	args := pgx.NamedArgs{
		"id":            t.ID,
		"number":        t.Number,
		"qr_token":      t.QRToken,
		"session_token": t.SessionToken,
		"status":        t.Status,
		"waiter_id":     t.WaiterID,
		"order_id":      t.OrderID,
	}

	cmdTag, err := r.db.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("table repository update: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return model.ErrNotFound
	}

	return nil
}

func (r *TableRepository) GetByID(ctx context.Context, id string) (*table.Table, error) {
	query := `
		SELECT id, venue_id, number, qr_token, session_token, status, waiter_id, order_id
		FROM tables
		WHERE id = $1
	`

	var t table.Table

	err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID,
		&t.VenueID,
		&t.Number,
		&t.QRToken,
		&t.SessionToken,
		&t.Status,
		&t.WaiterID,
		&t.OrderID,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("table repository get by id: %w", err)
	}

	return &t, nil
}

func (r *TableRepository) GetByFilters(ctx context.Context, filters *table.TableFilters) ([]*table.Table, error) {
	if filters == nil {
		return nil, ErrNoFilters
	}

	query, args, err := r.buildGetRequest(filters)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("table repository get by filters: %w", err)
	}
	defer rows.Close()

	var result []*table.Table

	for rows.Next() {
		var t table.Table

		if err := rows.Scan(
			&t.ID,
			&t.VenueID,
			&t.Number,
			&t.QRToken,
			&t.SessionToken,
			&t.Status,
			&t.WaiterID,
			&t.OrderID,
		); err != nil {
			return nil, fmt.Errorf("scan table: %w", err)
		}

		result = append(result, &t)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("rows error: %w", rows.Err())
	}

	return result, nil
}

func (r *TableRepository) buildGetRequest(filters *table.TableFilters) (string, []interface{}, error) {
	builder := squirrel.
		Select("id", "venue_id", "number", "qr_token", "session_token", "status", "waiter_id", "order_id").
		From("tables").
		PlaceholderFormat(squirrel.Dollar).
		OrderBy("number ASC")

	if filters.TableID != nil {
		builder = builder.Where(squirrel.Eq{"id": *filters.TableID})
	}
	if filters.QRToken != nil {
		builder = builder.Where(squirrel.Eq{"qr_token": *filters.QRToken})
	}
	if filters.WaiterID != nil {
		builder = builder.Where(squirrel.Eq{"waiter_id": *filters.WaiterID})
	}
	if filters.VenueID != nil {
		builder = builder.Where(squirrel.Eq{"venue_id": *filters.VenueID})
	}
	if filters.Session != nil {
		builder = builder.Where(squirrel.Eq{"session_token": *filters.Session})
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return "", nil, fmt.Errorf("%w: %w", ErrBuildQuery, err)
	}

	return query, args, nil
}

func (r *TableRepository) Delete(ctx context.Context, tableID string) error {
	query := `DELETE FROM tables WHERE id = $1`

	cmdTag, err := r.db.Exec(ctx, query, tableID)
	if err != nil {
		return fmt.Errorf("table repository delete: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return model.ErrNotFound
	}

	return nil
}
