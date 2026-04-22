package jwttokens

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tiptop-co/backend/internal/config"
	"github.com/tiptop-co/backend/internal/model/auth"
	"github.com/tiptop-co/backend/internal/model/authz"
)

func TestCreateAccessToken(t *testing.T) {
	cfg := config.AccessTokenConfig{
		TTL:       time.Minute * 15,
		Jitter:    time.Second * 5,
		SecretKey: "supersecretkey",
	}
	tokenService := New(cfg, config.RefreshTokenConfig{})

	claims := &auth.Claims{
		UserID:   uuid.New().String(),
		UserRole: authz.UserRole(1),
		VenueID:  uuid.New().String(),
	}

	token, err := tokenService.GenerateAccessToken(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	parsedClaims, err := tokenService.ParseAccessToken(token)
	assert.NoError(t, err)
	assert.Equal(t, claims.UserID, parsedClaims.UserID)
	assert.Equal(t, claims.UserRole, parsedClaims.UserRole)
	assert.Equal(t, claims.VenueID, parsedClaims.VenueID)
}

func TestTokenExpired(t *testing.T) {
	cfg := config.AccessTokenConfig{
		TTL:       time.Second,
		Jitter:    time.Second * 2,
		SecretKey: "supersecretkey",
	}
	tokenService := New(cfg, config.RefreshTokenConfig{})

	claims := &auth.Claims{
		UserID:   uuid.New().String(),
		UserRole: authz.UserRole(1),
		VenueID:  uuid.New().String(),
	}

	token, err := tokenService.GenerateAccessToken(claims)
	assert.NoError(t, err)

	time.Sleep(time.Second * 3)

	_, err = tokenService.ParseAccessToken(token)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrExpired))
}

func TestTokenParsingError(t *testing.T) {
	cfg := config.AccessTokenConfig{
		TTL:       time.Minute * 15,
		Jitter:    time.Second * 5,
		SecretKey: "supersecretkey",
	}
	tokenService := New(cfg, config.RefreshTokenConfig{})

	_, err := tokenService.ParseAccessToken("invalid.token.string")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrTokenParsing))
}

func TestTokenFromDifferentSource(t *testing.T) {
	cfg := config.AccessTokenConfig{
		TTL:       time.Minute * 15,
		Jitter:    time.Second * 5,
		SecretKey: "supersecretkey",
	}
	tokenService := New(cfg, config.RefreshTokenConfig{})

	claims := &auth.Claims{
		UserID:   uuid.New().String(),
		UserRole: authz.UserRole(1),
		VenueID:  uuid.New().String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims{
		UserID:   claims.UserID,
		UserRole: claims.UserRole,
		VenueID:  claims.VenueID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.TTL)),
		},
	})

	otherSecretKey := []byte("anothersecretkey")
	tokenStr, err := token.SignedString(otherSecretKey)
	assert.NoError(t, err)

	_, err = tokenService.ParseAccessToken(tokenStr)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrTokenParsing))
}
