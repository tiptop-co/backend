package waiter_request_postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tiptop-co/backend/internal/model"
	wrmodel "github.com/tiptop-co/backend/internal/model/waiter_request"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, req *wrmodel.Request) error {
	const query = `
		INSERT INTO waiter_requests (id, table_id, venue_id, waiter_id, status, created_at)
		VALUES (@id, @table_id, @venue_id, @waiter_id, @status, @created_at)
	`
	args := pgx.NamedArgs{
		"id":         req.ID,
		"table_id":   req.TableID,
		"venue_id":   req.VenueID,
		"waiter_id":  req.WaiterID,
		"status":     string(req.Status),
		"created_at": req.CreatedAt,
	}
	if _, err := r.db.Exec(ctx, query, args); err != nil {
		return fmt.Errorf("waiter_request repository create: %w", err)
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*wrmodel.Request, error) {
	const query = `
		SELECT wr.id, wr.table_id, COALESCE(t.number, 0), wr.venue_id, wr.waiter_id, wr.status, wr.created_at
		FROM waiter_requests wr
		LEFT JOIN tables t ON t.id = wr.table_id
		WHERE wr.id = $1
	`
	var (
		req       wrmodel.Request
		statusStr string
	)
	err := r.db.QueryRow(ctx, query, id).Scan(
		&req.ID, &req.TableID, &req.TableNumber, &req.VenueID, &req.WaiterID, &statusStr, &req.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("waiter_request repository get by id: %w", err)
	}
	req.Status = wrmodel.Status(statusStr)
	return &req, nil
}

func (r *Repository) GetByFilters(ctx context.Context, f *wrmodel.Filters) ([]*wrmodel.Request, error) {
	if f == nil {
		f = &wrmodel.Filters{}
	}

	builder := squirrel.
		Select("wr.id", "wr.table_id", "COALESCE(t.number, 0)", "wr.venue_id", "wr.waiter_id", "wr.status", "wr.created_at").
		From("waiter_requests wr").
		LeftJoin("tables t ON t.id = wr.table_id").
		PlaceholderFormat(squirrel.Dollar).
		OrderBy("wr.created_at ASC")

	if f.TableID != nil {
		builder = builder.Where(squirrel.Eq{"wr.table_id": *f.TableID})
	}
	if f.WaiterID != nil {
		builder = builder.Where(squirrel.Eq{"wr.waiter_id": *f.WaiterID})
	}
	if f.VenueID != nil {
		builder = builder.Where(squirrel.Eq{"wr.venue_id": *f.VenueID})
	}
	if len(f.Statuses) > 0 {
		strs := make([]string, 0, len(f.Statuses))
		for _, s := range f.Statuses {
			strs = append(strs, string(s))
		}
		builder = builder.Where(squirrel.Eq{"wr.status": strs})
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("waiter_request repository build query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("waiter_request repository query: %w", err)
	}
	defer rows.Close()

	var result []*wrmodel.Request
	for rows.Next() {
		var (
			req       wrmodel.Request
			statusStr string
		)
		if err := rows.Scan(&req.ID, &req.TableID, &req.TableNumber, &req.VenueID, &req.WaiterID, &statusStr, &req.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan waiter_request: %w", err)
		}
		req.Status = wrmodel.Status(statusStr)
		result = append(result, &req)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("rows: %w", rows.Err())
	}
	return result, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id string, status wrmodel.Status) error {
	const query = `UPDATE waiter_requests SET status = $1 WHERE id = $2`
	cmd, err := r.db.Exec(ctx, query, string(status), id)
	if err != nil {
		return fmt.Errorf("waiter_request repository update status: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *Repository) UpdateWaiter(ctx context.Context, id string, waiterID string) error {
	const query = `UPDATE waiter_requests SET waiter_id = $1 WHERE id = $2`
	cmd, err := r.db.Exec(ctx, query, waiterID, id)
	if err != nil {
		return fmt.Errorf("waiter_request repository update waiter: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}
