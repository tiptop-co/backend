package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/tiptop-co/backend/internal/model/auth"
	"github.com/tiptop-co/backend/internal/model/authz"
	"github.com/tiptop-co/backend/internal/model/table"
	"github.com/tiptop-co/backend/internal/providers/http/ctxclaims"
)

type TableSessionValidator interface {
	ValidateSessionToken(ctx context.Context, token string) (*table.Table, error)
}

func ParseTableSession(validator TableSessionValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := ctxclaims.GetClaims(c)
		// the user is already logged in as not a guest
		if claims != nil && claims.UserRole != authz.RoleGuest && claims.UserRole != authz.RoleUnspecified {
			c.Next()
			return
		}

		tableSession, err := c.Cookie("table_session")
		if err != nil {
			c.Next()
			return
		}

		table, err := validator.ValidateSessionToken(c.Request.Context(), tableSession)
		if err != nil {
			c.Next()
			return
		}

		c.Set("table", table)
		c.Set("claims", &auth.Claims{
			UserRole: authz.RoleGuest,
			VenueID:  table.VenueID,
		})

		c.Next()
	}
}
