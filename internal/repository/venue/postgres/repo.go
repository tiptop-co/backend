package venue_postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tiptop-co/backend/internal/model"
	"github.com/tiptop-co/backend/internal/model/venue"
)

type VenueRepository struct {
	db *pgxpool.Pool
}

func NewVenueRepository(db *pgxpool.Pool) *VenueRepository {
	return &VenueRepository{db: db}
}

func (r *VenueRepository) GetAll(ctx context.Context) ([]*venue.Venue, error) {
	const query = `
		SELECT id, name, address, description, bank_account, manager_id
		FROM venues
		ORDER BY name
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("venue repository get all: %w", err)
	}
	defer rows.Close()

	var res []*venue.Venue
	for rows.Next() {
		var v venue.Venue
		if err := rows.Scan(&v.ID, &v.Name, &v.Address, &v.Description, &v.BankAccount, &v.ManagerID); err != nil {
			return nil, fmt.Errorf("venue repository get all scan: %w", err)
		}
		res = append(res, &v)
	}
	return res, nil
}

func (r *VenueRepository) AssignManager(ctx context.Context, venueID, managerID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("venue repository assign manager begin: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			fmt.Printf("err : %w", err)
		}
	}()

	if _, err := tx.Exec(ctx,
		`UPDATE users SET venue_id = $1 WHERE id = $2`,
		venueID, managerID,
	); err != nil {
		return fmt.Errorf("venue repository attach manager: %w", err)
	}

	cmd, err := tx.Exec(ctx,
		`UPDATE venues SET manager_id = $1 WHERE id = $2`,
		managerID, venueID,
	)
	if err != nil {
		return fmt.Errorf("venue repository assign manager: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return tx.Commit(ctx)
}

func (r *VenueRepository) GetByID(ctx context.Context, id string) (*venue.Venue, error) {
	const query = `
		SELECT id, name, address, description, bank_account, manager_id
		FROM venues WHERE id = $1
	`
	var v venue.Venue
	err := r.db.QueryRow(ctx, query, id).Scan(
		&v.ID, &v.Name, &v.Address, &v.Description, &v.BankAccount, &v.ManagerID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("venue repository get by id: %w", err)
	}
	return &v, nil
}

func (r *VenueRepository) GetByManager(ctx context.Context, managerID string) (*venue.Venue, error) {
	const query = `
		SELECT id, name, address, description, bank_account, manager_id
		FROM venues WHERE manager_id = $1
		LIMIT 1
	`
	var v venue.Venue
	err := r.db.QueryRow(ctx, query, managerID).Scan(
		&v.ID, &v.Name, &v.Address, &v.Description, &v.BankAccount, &v.ManagerID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("venue repository get by manager: %w", err)
	}
	return &v, nil
}

func (r *VenueRepository) Update(ctx context.Context, v *venue.Venue) error {
	const query = `
		UPDATE venues
		SET name = @name, address = @address, description = @description, bank_account = @bank_account
		WHERE id = @id
	`
	args := pgx.NamedArgs{
		"id":           v.ID,
		"name":         v.Name,
		"address":      v.Address,
		"description":  v.Description,
		"bank_account": v.BankAccount,
	}
	cmd, err := r.db.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("venue repository update: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}
