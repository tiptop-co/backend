package admin_handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	statsusecase "github.com/tiptop-co/backend/internal/usecase/stats"
)

type StatsHandler struct {
	usecase statsusecase.Usecase
}

func NewStatsHandler(u statsusecase.Usecase) *StatsHandler {
	return &StatsHandler{usecase: u}
}

func (h *StatsHandler) Get(c *gin.Context) {
	out, err := h.usecase.GlobalStats(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, out)
}
