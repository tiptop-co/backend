package auth

import (
	"context"
	"errors"

	"github.com/tiptop-co/backend/internal/model/auth"
)

var (
	ErrWrongOldPassword   = errors.New("wrong old password")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type AuthUsecase interface {
	Login(ctx context.Context, credentials *auth.UserCredentials) (*auth.AuthTokens, error)
	Logout(ctx context.Context, refreshToken string) error
	RefreshTokens(ctx context.Context, refreshToken string) (*auth.AuthTokens, error)
	UpdatePassword(ctx context.Context, userID string, passwords *auth.UserPasswords) error
	GetClaims(ctx context.Context, accessToken string) (*auth.Claims, error)
}

type TokenService interface {
	GenerateTokens(claims *auth.Claims) (*auth.AuthTokens, error)
	ParseAccessToken(tokenStr string) (*auth.Claims, error)
}

type PasswordHasher interface {
	GenerateFromPassword(password string) (string, error)
	CompareHashAndPassword(hashed string, password string) error
}

type RefreshTokenRepository interface {
	Set(ctx context.Context, token string, claims *auth.Claims) error
	Get(ctx context.Context, token string) (*auth.Claims, error)
	Delete(ctx context.Context, token string) error
}

type AuthRepository interface {
	GetPasswordByID(ctx context.Context, userID string) (string, error)
	GetAuthInfoByLogin(ctx context.Context, login string) (*auth.AuthInfo, error)
	UpdatePassword(ctx context.Context, userID string, newPassword string) error
}
