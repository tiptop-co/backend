package ctxclaims

import (
	"github.com/gin-gonic/gin"
	"github.com/tiptop-co/backend/internal/model/auth"
)

func GetClaims(c *gin.Context) *auth.Claims {
	claims, exists := c.Get("claims")
	if !exists {
		return nil
	}

	parsedClaims, ok := claims.(*auth.Claims)
	if !ok {
		return nil
	}

	return parsedClaims
}
