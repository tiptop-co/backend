package table

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/tiptop-co/backend/internal/model"
	"github.com/tiptop-co/backend/internal/model/table"
)

var (
	ErrBuildQuery = errors.New("failed to build query")
	ErrNoFilters  = errors.New("filters must be in the request")
)

type TableRepository struct {
	db *sqlx.DB
}

func NewTableRepository(db *sqlx.DB) *TableRepository {
	return &TableRepository{
		db: db,
	}
}

func (r *TableRepository) Create(ctx context.Context, table table.Table) error {
	query := `
		INSERT INTO tables (id, venue_id, number, qr_token, session_token, status, waiter_id)
		VALUES (:id, :venue_id, :number, :qr_token, :session_token, :status, :waiter_id)
	`

	_, err := r.db.NamedExecContext(ctx, query, table)
	if err != nil {
		return fmt.Errorf("table repository create: %w", err)
	}

	return nil
}

func (r *TableRepository) Update(ctx context.Context, table table.Table) error {
	query := `
		UPDATE tables
		SET
			number        = :number,
			qr_token      = :qr_token,
			session_token = :session_token,
			status        = :status,
			waiter_id     = :waiter_id
		WHERE id = :id
	`

	_, err := r.db.NamedExecContext(ctx, query, table)
	if err != nil {
		return fmt.Errorf("table repository update: %w", err)
	}

	return nil
}

func (r *TableRepository) GetByID(ctx context.Context, id string) (*table.Table, error) {
	query := `
		SELECT id, venue_id, number, qr_token, session_token, status, waiter_id, created_at, updated_at
		FROM tables
		WHERE id = $1
	`
	var table table.Table

	err := r.db.GetContext(ctx, &table, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("table repository get by id: %w", err)
	}

	return &table, nil
}

func (r *TableRepository) GetByFilters(ctx context.Context, filters *table.TableFilters) ([]*table.Table, error) {
	builder := squirrel.Select("id", "venue_id", "number",
		"qr_token", "session_token", "status", "waiter_id").
		From("tables").PlaceholderFormat(squirrel.Dollar)
	if filters == nil {
		return nil, ErrNoFilters
	}

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

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBuildQuery, err)
	}

	var tables []*table.Table

	if err = r.db.SelectContext(ctx, &tables, query, args...); err != nil {
		return nil, fmt.Errorf("table repository get by filters: %w", err)
	}

	return tables, nil
}

func (r *TableRepository) Delete(ctx context.Context, tableID string) error {
	query := `DELETE FROM tables WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, tableID)
	if err != nil {
		return fmt.Errorf("table repository delete: %w", err)
	}

	return nil
}
