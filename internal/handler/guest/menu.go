package guest_handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tiptop-co/backend/internal/model"
	"github.com/tiptop-co/backend/internal/providers/http/middleware"
	"github.com/tiptop-co/backend/internal/usecase/menu"
	"github.com/tiptop-co/backend/pkg/errwrap"
)

type MenuHandler struct {
	usecase menu.MenuUsecase
}

func NewMenuHandler(u menu.MenuUsecase) *MenuHandler {
	return &MenuHandler{usecase: u}
}

func (h *MenuHandler) GetMenu(c *gin.Context) {
	t := middleware.GetTable(c)
	if t == nil {
		_ = c.Error(model.ErrUnauthorized)
		return
	}

	tableID := c.Param("table_id")
	if tableID == "" || tableID != t.ID {
		_ = c.Error(errwrap.Wrap(model.ErrValidation, model.ErrForbidden))
		return
	}

	dishes, cats, err := h.usecase.GetVenueMenu(c.Request.Context(), t.VenueID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, &getMenuResponse{Dishes: dishes, Categories: cats})
}

func (h *MenuHandler) GetDish(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		_ = c.Error(model.ErrValidation)
		return
	}

	d, err := h.usecase.GetDish(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, d)
}
