package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tiptop-co/backend/internal/model/auth"
	"github.com/tiptop-co/backend/internal/providers/http/cookie"
	jwttokens "github.com/tiptop-co/backend/internal/providers/tokens/jwt"
)

type AccessTokenParser interface {
	GetClaims(ctx context.Context, token string) (*auth.Claims, error)
}

func ClaimsRequired(parser AccessTokenParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(cookie.AccessCookieName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		claims, err := parser.GetClaims(c, token)
		if err != nil {
			var h gin.H
			if errors.Is(err, jwttokens.ErrExpired) {
				h = gin.H{"error": "token expired"}
			} else {
				h = gin.H{"error": "invalid token"}
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, h)
			return
		}

		c.Set("claims", claims)
		c.Next()
	}
}

func ClaimsOptional(parser AccessTokenParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(cookie.AccessCookieName)
		if err != nil {
			c.Next()
		}

		claims, err := parser.GetClaims(c, token)
		if err != nil {
			c.Next()
		}

		c.Set("claims", claims)
		c.Next()
	}
}
