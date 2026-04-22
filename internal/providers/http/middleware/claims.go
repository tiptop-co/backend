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
