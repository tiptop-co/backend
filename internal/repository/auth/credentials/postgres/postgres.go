package creds_postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tiptop-co/backend/internal/model/auth"
)

type CredsRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *CredsRepository {
	return &CredsRepository{pool: pool}
}

func (r *CredsRepository) SaveCredentials(ctx context.Context, userID uuid.UUID, credentials *auth.UserCredentials) error {
	const query = `
		INSERT INTO credentials (user_id, login, password)
		VALUES ($1, $2, $3)
		ON CONFLICT (login) DO NOTHING
	`

	_, err := r.pool.Exec(ctx, query, userID, credentials.Login, credentials.Password)
	if err != nil {
		return fmt.Errorf("save credentials: %w", err)
	}

	return nil
}

func (r *CredsRepository) GetPasswordByID(ctx context.Context, userID string) (string, error) {
	const query = `
		SELECT password
		FROM credentials
		WHERE user_id = $1
	`

	var password string
	err := r.pool.QueryRow(ctx, query, userID).Scan(&password)
	if err != nil {
		return "", fmt.Errorf("get password by id: %w", err)
	}

	return password, nil
}

func (r *CredsRepository) GetAuthInfoByLogin(ctx context.Context, login string) (*auth.AuthInfo, error) {
	const query = `
		SELECT 
			c.password,
			u.id,
			u.role,
			u.venue_id
		FROM credentials c
		JOIN users u ON u.id = c.user_id
		WHERE c.login = $1
	`

	var (
		password string
		claims   auth.Claims
	)

	err := r.pool.QueryRow(ctx, query, login).Scan(
		&password,
		&claims.UserID,
		&claims.UserRole,
		&claims.VenueID,
	)
	if err != nil {
		return nil, fmt.Errorf("get auth info by login: %w", err)
	}

	return &auth.AuthInfo{
		HashedPassword: password,
		Claims:         &claims,
	}, nil
}

func (r *CredsRepository) UpdatePassword(ctx context.Context, userID string, newPassword string) error {
	const query = `
		UPDATE credentials
		SET password = $1
		WHERE user_id = $2
	`

	ct, err := r.pool.Exec(ctx, query, newPassword, userID)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("user %s not found", userID)
	}

	return nil
}
