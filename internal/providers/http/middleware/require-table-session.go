package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/tiptop-co/backend/internal/model"
	"github.com/tiptop-co/backend/internal/model/table"
)

func RequireTableSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get("table")
		if !ok {
			_ = c.AbortWithError(401, model.ErrUnauthorized)
			return
		}
		if _, ok := v.(*table.Table); !ok {
			_ = c.AbortWithError(401, model.ErrUnauthorized)
			return
		}
		c.Next()
	}
}

func GetTable(c *gin.Context) *table.Table {
	v, ok := c.Get("table")
	if !ok {
		return nil
	}
	t, _ := v.(*table.Table)
	return t
}
