package jwttokens

import (
	cryptorand "crypto/rand"
	"encoding/base64"
	"errors"
	mathrand "math/rand"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tiptop-co/backend/internal/config"
	"github.com/tiptop-co/backend/internal/model/auth"
)

var (
	ErrExpired              = errors.New("access token expired error")
	ErrInvalidClaims        = errors.New("invalid claims error")
	ErrInvalidSigningMethod = errors.New("invalid signing method error")
	ErrTokenParsing         = errors.New("token parsing error")
	ErrInvalidAccessToken   = errors.New("invalid access token")
)

type jwtClaims struct {
	UserID   string        `json:"user_id"`
	UserRole auth.UserRole `json:"user_role"`
	VenueID  string        `json:"venue_id"`
	jwt.RegisteredClaims
}

type (
	accessCfg  = config.AccessTokenConfig
	refreshCfg = config.RefreshTokenConfig
)

type JWTTokenService struct {
	accessCfg accessCfg
	refresh   refreshCfg
}

func New(access accessCfg, refresh refreshCfg) *JWTTokenService {
	return &JWTTokenService{
		accessCfg: access,
		refresh:   refresh,
	}
}

func (s *JWTTokenService) GenerateAccessToken(claims *auth.Claims) (string, error) {
	jitter := time.Duration(mathrand.Int63n(int64(s.accessCfg.Jitter)))
	exp := time.Now().Add(s.accessCfg.TTL + jitter)

	jwtClaims := jwtClaims{
		UserID:   claims.UserID,
		UserRole: claims.UserRole,
		VenueID:  claims.VenueID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)

	return token.SignedString(s.accessCfg.SecretKey)
}

func (s *JWTTokenService) ParseAccessToken(tokenStr string) (*auth.Claims, error) {
	var jwtClaims jwtClaims

	token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSigningMethod
		}
		return s.accessCfg.SecretKey, nil
	})

	switch {
	case errors.Is(err, jwt.ErrTokenExpired):
		return nil, ErrExpired
	case errors.Is(err, jwt.ErrTokenInvalidClaims):
		return nil, ErrInvalidClaims
	case err != nil:
		return nil, ErrTokenParsing
	case !token.Valid:
		return nil, ErrInvalidAccessToken
	}

	return &auth.Claims{
		UserID:   jwtClaims.UserID,
		UserRole: jwtClaims.UserRole,
		VenueID:  jwtClaims.VenueID,
	}, nil
}

func (s *JWTTokenService) GenerateRefreshToken() (string, error) {
	b := make([]byte, s.refresh.Size)

	_, err := cryptorand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (s *JWTTokenService) GenerateTokens(claims *auth.Claims) (*auth.AuthTokens, error) {
	access, err := s.GenerateAccessToken(claims)
	if err != nil {
		return nil, err
	}

	refresh, err := s.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	return &auth.AuthTokens{
		Access:  access,
		Refresh: refresh,
	}, nil
}
