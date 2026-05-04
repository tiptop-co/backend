package user_postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tiptop-co/backend/internal/model"
	"github.com/tiptop-co/backend/internal/model/authz"
	"github.com/tiptop-co/backend/internal/model/user"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

const baseSelect = `
	SELECT u.id, u.first_name, u.last_name, COALESCE(c.login,'') AS phone,
	       u.role, u.venue_id, v.name AS venue_name, u.created_at
	FROM users u
	LEFT JOIN credentials c ON c.user_id = u.id
	LEFT JOIN venues v ON v.id = u.venue_id
`

func scan(row pgx.Row) (*user.User, error) {
	var (
		u         user.User
		role      int
		venueID   *string
		venueName *string
	)
	err := row.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Phone, &role, &venueID, &venueName, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	u.Role = authz.UserRole(role)
	u.VenueID = venueID
	u.VenueName = venueName
	return &u, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	row := r.db.QueryRow(ctx, baseSelect+` WHERE u.id = $1`, id)
	u, err := scan(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user repository get by id: %w", err)
	}
	return u, nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, id, firstName, lastName string) error {
	const query = `UPDATE users SET first_name = $1, last_name = $2 WHERE id = $3`
	cmd, err := r.db.Exec(ctx, query, firstName, lastName, id)
	if err != nil {
		return fmt.Errorf("user repository update profile: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *UserRepository) GetByVenueAndRole(ctx context.Context, venueID string, role authz.UserRole) ([]*user.User, error) {
	rows, err := r.db.Query(ctx, baseSelect+` WHERE u.venue_id = $1 AND u.role = $2 ORDER BY u.created_at`, venueID, int(role))
	if err != nil {
		return nil, fmt.Errorf("user repository list by venue+role: %w", err)
	}
	defer rows.Close()

	var result []*user.User
	for rows.Next() {
		u, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		result = append(result, u)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("rows: %w", rows.Err())
	}
	return result, nil
}

func (r *UserRepository) Create(ctx context.Context, u *user.User, login, hashedPassword string) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("user repository begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	const insertUser = `
		INSERT INTO users (id, first_name, last_name, role, venue_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	if _, err = tx.Exec(ctx, insertUser, u.ID, u.FirstName, u.LastName, int(u.Role), u.VenueID, u.CreatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			err = fmt.Errorf("%w: %w", model.ErrAlreadyExists, err)
		}
		return fmt.Errorf("user repository insert user: %w", err)
	}

	const insertCreds = `INSERT INTO credentials (user_id, login, password) VALUES ($1, $2, $3)`
	if _, err = tx.Exec(ctx, insertCreds, u.ID, login, hashedPassword); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			err = fmt.Errorf("%w: %w", model.ErrAlreadyExists, err)
		}
		return fmt.Errorf("user repository insert credentials: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("user repository commit: %w", err)
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("user repository begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	if _, err = tx.Exec(ctx, `DELETE FROM credentials WHERE user_id = $1`, id); err != nil {
		return fmt.Errorf("user repository delete credentials: %w", err)
	}
	cmd, err := tx.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("user repository delete user: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		err = model.ErrNotFound
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("user repository commit: %w", err)
	}
	return nil
}

func (r *UserRepository) DeleteScoped(ctx context.Context, id, venueID string, role authz.UserRole) error {
	const checkQuery = `SELECT 1 FROM users WHERE id = $1 AND venue_id = $2 AND role = $3`
	var x int
	err := r.db.QueryRow(ctx, checkQuery, id, venueID, int(role)).Scan(&x)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("user repository scoped check: %w", err)
	}
	return r.Delete(ctx, id)
}

type WaiterStats struct {
	User              *user.User
	TablesServedToday int
	TipsToday         int
}

func (r *UserRepository) GetWaiterStatsByVenue(ctx context.Context, venueID string) ([]*WaiterStats, error) {
	now := time.Now().UTC()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.Add(24 * time.Hour)

	const query = `
		SELECT u.id, u.first_name, u.last_name, COALESCE(c.login,'') AS phone,
		       u.role, u.venue_id, v.name AS venue_name, u.created_at,
		       COALESCE((
		           SELECT COUNT(DISTINCT o.table_id) FROM orders o
		           WHERE o.waiter_id = u.id
		             AND o.created_at >= $2 AND o.created_at < $3
		       ), 0) AS tables_today,
		       COALESCE((
		           SELECT SUM(t.tips_amount) FROM transactions t
		           JOIN orders o2 ON o2.id = t.order_id
		           WHERE o2.waiter_id = u.id AND t.status = 'success'
		             AND t.created_at >= $2 AND t.created_at < $3
		       ), 0) AS tips_today
		FROM users u
		LEFT JOIN credentials c ON c.user_id = u.id
		LEFT JOIN venues v ON v.id = u.venue_id
		WHERE u.venue_id = $1 AND u.role = $4
		ORDER BY u.created_at
	`
	rows, err := r.db.Query(ctx, query, venueID, dayStart, dayEnd, int(authz.RoleWaiter))
	if err != nil {
		return nil, fmt.Errorf("user repository waiter stats: %w", err)
	}
	defer rows.Close()

	var result []*WaiterStats
	for rows.Next() {
		var (
			u             user.User
			role          int
			venueIDPtr    *string
			venueName     *string
			tables, tips  int
		)
		if err := rows.Scan(
			&u.ID, &u.FirstName, &u.LastName, &u.Phone, &role,
			&venueIDPtr, &venueName, &u.CreatedAt,
			&tables, &tips,
		); err != nil {
			return nil, fmt.Errorf("scan waiter stats: %w", err)
		}
		u.Role = authz.UserRole(role)
		u.VenueID = venueIDPtr
		u.VenueName = venueName
		result = append(result, &WaiterStats{User: &u, TablesServedToday: tables, TipsToday: tips})
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("rows: %w", rows.Err())
	}
	return result, nil
}

func (r *UserRepository) GetByRole(ctx context.Context, role authz.UserRole) ([]*user.User, error) {
	rows, err := r.db.Query(ctx, baseSelect+` WHERE u.role = $1 ORDER BY u.created_at`, int(role))
	if err != nil {
		return nil, fmt.Errorf("user repository list by role: %w", err)
	}
	defer rows.Close()

	var result []*user.User
	for rows.Next() {
		u, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		result = append(result, u)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("rows: %w", rows.Err())
	}
	return result, nil
}
