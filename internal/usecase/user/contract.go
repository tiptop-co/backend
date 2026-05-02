package user

import (
	"context"
	"errors"

	"github.com/tiptop-co/backend/internal/model/authz"
	usermodel "github.com/tiptop-co/backend/internal/model/user"
	user_postgres "github.com/tiptop-co/backend/internal/repository/user/postgres"
)

var (
	ErrInvalidPhone = errors.New("invalid phone format")
)

type CreateAccountResult struct {
	User              *usermodel.User `json:"user"`
	GeneratedPassword string          `json:"generated_password"`
}

type WaiterUsecase interface {
	CreateWaiter(ctx context.Context, venueID, firstName, lastName, phone string) (*CreateAccountResult, error)
	DeleteWaiter(ctx context.Context, venueID, userID string) error
	GetWaitersWithStats(ctx context.Context, venueID string) ([]*user_postgres.WaiterStats, error)
}

type ManagerUsecase interface {
	CreateManager(ctx context.Context, venueID *string, firstName, lastName, phone string) (*CreateAccountResult, error)
	DeleteManager(ctx context.Context, userID string) error
	GetManagers(ctx context.Context) ([]*usermodel.User, error)
}

type Repository interface {
	Create(ctx context.Context, u *usermodel.User, login, hashedPassword string) error
	Delete(ctx context.Context, id string) error
	DeleteScoped(ctx context.Context, id, venueID string, role authz.UserRole) error
	GetByID(ctx context.Context, id string) (*usermodel.User, error)
	GetByRole(ctx context.Context, role authz.UserRole) ([]*usermodel.User, error)
	GetWaiterStatsByVenue(ctx context.Context, venueID string) ([]*user_postgres.WaiterStats, error)
}

type PasswordHasher interface {
	GenerateFromPassword(password string) (string, error)
}

type PasswordGenerator interface {
	Generate() (string, error)
}
