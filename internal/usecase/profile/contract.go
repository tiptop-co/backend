package profile

import (
	"context"

	"github.com/tiptop-co/backend/internal/model/auth"
	"github.com/tiptop-co/backend/internal/model/user"
)

type ProfileUsecase interface {
	GetMe(ctx context.Context, userID string) (*user.User, error)
	UpdateProfile(ctx context.Context, userID, firstName, lastName string) (*user.User, error)
	ChangePassword(ctx context.Context, userID string, current, new string) error
	GetTodayTips(ctx context.Context, waiterID string) (int, error)
}

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*user.User, error)
	UpdateProfile(ctx context.Context, id, firstName, lastName string) error
}

type TransactionRepository interface {
	SumTipsByWaiterToday(ctx context.Context, waiterID string) (int, error)
}

type PasswordChanger interface {
	UpdatePassword(ctx context.Context, userID string, passwords *auth.UserPasswords) error
}
