package profile

import (
	"context"

	"github.com/tiptop-co/backend/internal/model"
	"github.com/tiptop-co/backend/internal/model/auth"
	"github.com/tiptop-co/backend/internal/model/user"
	"github.com/tiptop-co/backend/pkg/errwrap"
)

const errPrefix = "[PROFILE USECASE]"

type ProfileService struct {
	users    UserRepository
	txRepo   TransactionRepository
	pwd      PasswordChanger
}

func NewProfileService(users UserRepository, tx TransactionRepository, pwd PasswordChanger) *ProfileService {
	return &ProfileService{users: users, txRepo: tx, pwd: pwd}
}

func (s *ProfileService) GetMe(ctx context.Context, userID string) (_ *user.User, err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" get me", err) }()

	if userID == "" {
		return nil, model.ErrValidation
	}
	return s.users.GetByID(ctx, userID)
}

func (s *ProfileService) UpdateProfile(ctx context.Context, userID, firstName, lastName string) (_ *user.User, err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" update profile", err) }()

	if userID == "" || firstName == "" || lastName == "" {
		return nil, model.ErrValidation
	}
	if err := s.users.UpdateProfile(ctx, userID, firstName, lastName); err != nil {
		return nil, err
	}
	return s.users.GetByID(ctx, userID)
}

func (s *ProfileService) ChangePassword(ctx context.Context, userID, current, newPwd string) (err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" change password", err) }()

	if userID == "" || current == "" || newPwd == "" {
		return model.ErrValidation
	}
	return s.pwd.UpdatePassword(ctx, userID, &auth.UserPasswords{Old: current, New: newPwd})
}

func (s *ProfileService) GetTodayTips(ctx context.Context, waiterID string) (_ int, err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" get today tips", err) }()

	if waiterID == "" {
		return 0, model.ErrValidation
	}
	return s.txRepo.SumTipsByWaiterToday(ctx, waiterID)
}
