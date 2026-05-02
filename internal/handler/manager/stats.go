package manager_handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tiptop-co/backend/internal/model"
	"github.com/tiptop-co/backend/internal/providers/http/ctxclaims"
	statsusecase "github.com/tiptop-co/backend/internal/usecase/stats"
)

type StatsHandler struct {
	usecase statsusecase.Usecase
}

func NewStatsHandler(u statsusecase.Usecase) *StatsHandler {
	return &StatsHandler{usecase: u}
}

func (h *StatsHandler) Get(c *gin.Context) {
	claims := ctxclaims.GetClaims(c)
	if claims == nil || claims.VenueID == "" {
		_ = c.Error(model.ErrUnauthorized)
		return
	}
	out, err := h.usecase.VenueStats(c.Request.Context(), claims.VenueID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, out)
}
