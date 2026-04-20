package auth

import (
	"context"

	"github.com/tiptop-co/backend/internal/model/auth"
	"github.com/tiptop-co/backend/pkg/errwrap"
)

const (
	errPrefix = "[AUTH USECASE]"
)

type AuthService struct {
	authRepo     AuthRepository
	refreshRepo  RefreshTokenRepository
	hasher       PasswordHasher
	tokenService TokenService
}

func NewAuthService(authRepo AuthRepository, refreshRepo RefreshTokenRepository,
	hasher PasswordHasher, token TokenService) *AuthService {
	return &AuthService{
		authRepo:     authRepo,
		refreshRepo:  refreshRepo,
		hasher:       hasher,
		tokenService: token,
	}
}

func (a *AuthService) Login(ctx context.Context, credentials *auth.UserCredentials) (_ *auth.AuthTokens, err error) {
	defer func() {
		err = errwrap.WrapMsg(errPrefix, err)
	}()

	info, err := a.authRepo.GetAuthInfoByLogin(ctx, credentials.Login)
	if err != nil {
		return nil, err
	}

	if a.hasher.CompareHashAndPassword(info.HashedPassword, credentials.Password) != nil {
		return nil, ErrInvalidCredentials
	}

	tokens, err := a.tokenService.GenerateTokens(info.Claims)
	if err != nil {
		return nil, err
	}

	err = a.refreshRepo.Set(ctx, tokens.Refresh, info.Claims)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (a *AuthService) Logout(ctx context.Context, refreshToken string) (err error) {
	defer func() {
		err = errwrap.WrapMsg(errPrefix, err)
	}()

	return a.refreshRepo.Delete(ctx, refreshToken)
}

func (a *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (_ *auth.AuthTokens, err error) {
	defer func() {
		err = errwrap.WrapMsg(errPrefix, err)
	}()

	claims, err := a.refreshRepo.Get(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	newTokens, err := a.tokenService.GenerateTokens(claims)
	if err != nil {
		return nil, err
	}

	err = a.refreshRepo.Delete(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	err = a.refreshRepo.Set(ctx, newTokens.Refresh, claims)
	if err != nil {
		return nil, err
	}

	return newTokens, nil
}

func (a *AuthService) UpdatePassword(ctx context.Context, userID string, passwords *auth.UserPasswords) (err error) {
	defer func() {
		err = errwrap.WrapMsg(errPrefix, err)
	}()

	hashedPassword, err := a.authRepo.GetPasswordByID(ctx, userID)
	if err != nil {
		return err
	}

	if a.hasher.CompareHashAndPassword(hashedPassword, passwords.Old) != nil {
		return ErrWrongOldPassword
	}

	newHashed, err := a.hasher.GenerateFromPassword(passwords.New)
	if err != nil {
		return err
	}

	return a.authRepo.UpdatePassword(ctx, userID, string(newHashed))
}

func (a *AuthService) GetClaims(ctx context.Context, accessToken string) (_ *auth.Claims, err error) {
	defer func() {
		err = errwrap.WrapMsg(errPrefix, err)
	}()

	return a.tokenService.ParseAccessToken(accessToken)
}
