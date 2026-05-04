package manager_handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tiptop-co/backend/internal/model"
	venuemodel "github.com/tiptop-co/backend/internal/model/venue"
	"github.com/tiptop-co/backend/internal/providers/http/ctxclaims"
	venueusecase "github.com/tiptop-co/backend/internal/usecase/venue"
	"github.com/tiptop-co/backend/pkg/errwrap"
)

type updateVenueRequest struct {
	Name        string  `json:"name"         binding:"required"`
	Address     string  `json:"address"      binding:"required"`
	Description *string `json:"description"`
	BankAccount *string `json:"bank_account"`
}

type VenueHandler struct {
	usecase venueusecase.VenueUsecase
}

func NewVenueHandler(u venueusecase.VenueUsecase) *VenueHandler {
	return &VenueHandler{usecase: u}
}

func (h *VenueHandler) Get(c *gin.Context) {
	claims := ctxclaims.GetClaims(c)
	if claims == nil {
		_ = c.Error(model.ErrUnauthorized)
		return
	}
	v, err := h.usecase.GetByManager(c.Request.Context(), claims.UserID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, v)
}

func (h *VenueHandler) Update(c *gin.Context) {
	claims := ctxclaims.GetClaims(c)
	if claims == nil {
		_ = c.Error(model.ErrUnauthorized)
		return
	}
	var req updateVenueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errwrap.Wrap(model.ErrValidation, err))
		return
	}
	v, err := h.usecase.UpdateByManager(c.Request.Context(), claims.UserID, &venuemodel.UpdateInput{
		Name:        req.Name,
		Address:     req.Address,
		Description: req.Description,
		BankAccount: req.BankAccount,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, v)
}
