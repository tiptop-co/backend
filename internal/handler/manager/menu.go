package manager_handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tiptop-co/backend/internal/model"
	"github.com/tiptop-co/backend/internal/model/dish"
	"github.com/tiptop-co/backend/internal/providers/http/ctxclaims"
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
	claims := ctxclaims.GetClaims(c)
	if claims == nil || claims.VenueID == "" {
		_ = c.Error(model.ErrUnauthorized)
		return
	}

	dishes, cats, err := h.usecase.GetVenueMenu(c.Request.Context(), claims.VenueID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, &getMenuResponse{Dishes: dishes, Categories: cats})
}

func (h *MenuHandler) CreateDish(c *gin.Context) {
	claims := ctxclaims.GetClaims(c)
	if claims == nil || claims.VenueID == "" {
		_ = c.Error(model.ErrUnauthorized)
		return
	}

	var req createDishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errwrap.Wrap(model.ErrValidation, err))
		return
	}

	in := &dish.CreateInput{
		Name:        req.Name,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		Price:       req.Price,
		Weight:      req.Weight,
		WeightUnit:  dish.WeightUnit(req.WeightUnit),
		Calories:    req.Calories,
		Protein:     req.Protein,
		Fat:         req.Fat,
		Carbs:       req.Carbs,
	}

	d, err := h.usecase.CreateDish(c.Request.Context(), claims.VenueID, in)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, d)
}

func (h *MenuHandler) DeleteDish(c *gin.Context) {
	claims := ctxclaims.GetClaims(c)
	if claims == nil || claims.VenueID == "" {
		_ = c.Error(model.ErrUnauthorized)
		return
	}

	id := c.Param("id")
	if id == "" {
		_ = c.Error(model.ErrValidation)
		return
	}

	if err := h.usecase.DeleteDish(c.Request.Context(), claims.VenueID, id); err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusOK)
}
