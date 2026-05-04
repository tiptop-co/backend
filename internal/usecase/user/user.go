package user

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/tiptop-co/backend/internal/model"
	"github.com/tiptop-co/backend/internal/model/authz"
	usermodel "github.com/tiptop-co/backend/internal/model/user"
	user_postgres "github.com/tiptop-co/backend/internal/repository/user/postgres"
	"github.com/tiptop-co/backend/pkg/errwrap"
	"github.com/tiptop-co/backend/pkg/phone"
)

const errPrefix = "[USER USECASE]"

type Service struct {
	repo   Repository
	hasher PasswordHasher
	gen    PasswordGenerator
}

func NewService(repo Repository, hasher PasswordHasher, gen PasswordGenerator) *Service {
	return &Service{repo: repo, hasher: hasher, gen: gen}
}

func (s *Service) CreateWaiter(ctx context.Context, venueID, firstName, lastName, phoneStr string) (_ *CreateAccountResult, err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" create waiter", err) }()
	return s.createAccount(ctx, &venueID, authz.RoleWaiter, firstName, lastName, phoneStr)
}

func (s *Service) DeleteWaiter(ctx context.Context, venueID, userID string) (err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" delete waiter", err) }()

	if venueID == "" || userID == "" {
		return model.ErrValidation
	}
	return s.repo.DeleteScoped(ctx, userID, venueID, authz.RoleWaiter)
}

func (s *Service) GetWaitersWithStats(ctx context.Context, venueID string) (_ []*user_postgres.WaiterStats, err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" get waiter stats", err) }()

	if venueID == "" {
		return nil, model.ErrValidation
	}
	return s.repo.GetWaiterStatsByVenue(ctx, venueID)
}

func (s *Service) CreateManager(ctx context.Context, venueID *string, firstName, lastName, phoneStr string) (_ *CreateAccountResult, err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" create manager", err) }()
	return s.createAccount(ctx, venueID, authz.RoleManager, firstName, lastName, phoneStr)
}

func (s *Service) DeleteManager(ctx context.Context, userID string) (err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" delete manager", err) }()

	if userID == "" {
		return model.ErrValidation
	}
	u, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if u.Role != authz.RoleManager {
		return model.ErrNotFound
	}
	return s.repo.Delete(ctx, userID)
}

func (s *Service) GetManagers(ctx context.Context) (_ []*usermodel.User, err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" get managers", err) }()
	return s.repo.GetByRole(ctx, authz.RoleManager)
}

func (s *Service) createAccount(ctx context.Context, venueID *string, role authz.UserRole, firstName, lastName, phoneStr string) (*CreateAccountResult, error) {
	if firstName == "" || lastName == "" {
		return nil, model.ErrValidation
	}
	if !phone.IsValid(phoneStr) {
		return nil, ErrInvalidPhone
	}

	rawPassword, err := s.gen.Generate()
	if err != nil {
		return nil, err
	}
	hashed, err := s.hasher.GenerateFromPassword(rawPassword)
	if err != nil {
		return nil, err
	}

	u := &usermodel.User{
		ID:        uuid.New().String(),
		FirstName: firstName,
		LastName:  lastName,
		Phone:     phoneStr,
		Role:      role,
		VenueID:   venueID,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, u, phoneStr, hashed); err != nil {
		return nil, err
	}

	created, err := s.repo.GetByID(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	return &CreateAccountResult{User: created, GeneratedPassword: rawPassword}, nil
}
