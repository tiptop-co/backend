package manager_handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tiptop-co/backend/internal/model"
	usermodel "github.com/tiptop-co/backend/internal/model/user"
	"github.com/tiptop-co/backend/internal/providers/http/ctxclaims"
	userusecase "github.com/tiptop-co/backend/internal/usecase/user"
	"github.com/tiptop-co/backend/pkg/errwrap"
)

type createWaiterRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name"  binding:"required"`
	Phone     string `json:"phone"      binding:"required"`
}

type waiterStatsItem struct {
	User              *usermodel.User `json:"user"`
	TablesServedToday int             `json:"tables_served_today"`
	TipsToday         int             `json:"tips_today"`
}

type waitersListResponse struct {
	Data []waiterStatsItem `json:"data"`
}

type WaitersHandler struct {
	usecase userusecase.WaiterUsecase
}

func NewWaitersHandler(u userusecase.WaiterUsecase) *WaitersHandler {
	return &WaitersHandler{usecase: u}
}

func (h *WaitersHandler) List(c *gin.Context) {
	claims := ctxclaims.GetClaims(c)
	if claims == nil || claims.VenueID == "" {
		_ = c.Error(model.ErrUnauthorized)
		return
	}
	stats, err := h.usecase.GetWaitersWithStats(c.Request.Context(), claims.VenueID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	out := make([]waiterStatsItem, 0, len(stats))
	for _, s := range stats {
		out = append(out, waiterStatsItem{User: s.User, TablesServedToday: s.TablesServedToday, TipsToday: s.TipsToday})
	}
	c.JSON(http.StatusOK, &waitersListResponse{Data: out})
}

func (h *WaitersHandler) Create(c *gin.Context) {
	claims := ctxclaims.GetClaims(c)
	if claims == nil || claims.VenueID == "" {
		_ = c.Error(model.ErrUnauthorized)
		return
	}
	var req createWaiterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errwrap.Wrap(model.ErrValidation, err))
		return
	}
	res, err := h.usecase.CreateWaiter(c.Request.Context(), claims.VenueID, req.FirstName, req.LastName, req.Phone)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, res)
}

func (h *WaitersHandler) Delete(c *gin.Context) {
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
	if err := h.usecase.DeleteWaiter(c.Request.Context(), claims.VenueID, id); err != nil {
		_ = c.Error(err)
		return
	}
	c.Status(http.StatusOK)
}
