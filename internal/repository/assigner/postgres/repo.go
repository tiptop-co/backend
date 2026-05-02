package assigner_postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tiptop-co/backend/internal/model/authz"
)

type AssignerRepository struct {
	db *pgxpool.Pool
}

func NewAssignerRepository(db *pgxpool.Pool) *AssignerRepository {
	return &AssignerRepository{db: db}
}

func (r *AssignerRepository) LeastLoadedWaiter(ctx context.Context, venueID string) (*string, error) {
	const query = `
		SELECT u.id
		FROM users u
		LEFT JOIN tables t ON t.waiter_id = u.id AND t.status != 'free'
		WHERE u.role = $1 AND u.venue_id = $2
		GROUP BY u.id
		ORDER BY COUNT(t.id) ASC, u.created_at ASC
		LIMIT 1
	`
	var id string
	err := r.db.QueryRow(ctx, query, int(authz.RoleWaiter), venueID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("assigner repository least loaded: %w", err)
	}
	return &id, nil
}
