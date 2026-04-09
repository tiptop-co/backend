package table

import (
	"context"

	"github.com/gin-gonic/gin"
)

type TableSessionValidator interface {
	ValidateSessionToken(ctx context.Context, tableID string, sessionToken string) (err error)
}

func TableSessionMiddleware(validator TableSessionValidator) gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}
