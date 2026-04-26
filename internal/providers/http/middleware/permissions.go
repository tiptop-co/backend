package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tiptop-co/backend/internal/model"
	"github.com/tiptop-co/backend/internal/model/authz"
	"github.com/tiptop-co/backend/internal/providers/http/ctxclaims"
)

func RequirePermission(p authz.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := ctxclaims.GetClaims(c)
		if claims == nil {
			_ = c.AbortWithError(http.StatusUnauthorized, model.ErrUnauthorized)
			return
		}

		if !authz.HasPermission(claims.UserRole, p) {
			_ = c.AbortWithError(http.StatusForbidden, model.ErrForbidden)
			return
		}

		c.Next()
	}
}
