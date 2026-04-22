package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/tiptop-co/backend/internal/model/auth"
	"github.com/tiptop-co/backend/internal/providers/http/cookie"
)

type Parser interface {
	GetClaims(ctx context.Context, token string) (*auth.Claims, error)
}

func ParseClaims(parser Parser) gin.HandlerFunc {
	return func(c *gin.Context) {
		if token, err := c.Cookie(cookie.AccessCookieName); err == nil {
			if claims, err := parser.GetClaims(c, token); err == nil {
				c.Set("claims", claims)
			}
		}

		c.Next()
	}
}
