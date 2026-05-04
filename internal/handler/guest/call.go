package guest_handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tiptop-co/backend/internal/model"
	wrmodel "github.com/tiptop-co/backend/internal/model/waiter_request"
	"github.com/tiptop-co/backend/internal/providers/http/middleware"
	wrusecase "github.com/tiptop-co/backend/internal/usecase/waiter_request"
	"github.com/tiptop-co/backend/pkg/errwrap"
)

type callStatusResponse struct {
	CanCall bool `json:"can_call"`
}

type tableRequestsResponse struct {
	Data []*wrmodel.Request `json:"data"`
}

type createCallRequest struct {
	TableID string `json:"table_id" binding:"required"`
}

type CallHandler struct {
	usecase wrusecase.Usecase
}

func NewCallHandler(u wrusecase.Usecase) *CallHandler {
	return &CallHandler{usecase: u}
}

func (h *CallHandler) Status(c *gin.Context) {
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
	can, err := h.usecase.CanCall(c.Request.Context(), tableID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, &callStatusResponse{CanCall: can})
}

func (h *CallHandler) List(c *gin.Context) {
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
	reqs, err := h.usecase.GetByTable(c.Request.Context(), tableID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, &tableRequestsResponse{Data: reqs})
}

func (h *CallHandler) Create(c *gin.Context) {
	t := middleware.GetTable(c)
	if t == nil {
		_ = c.Error(model.ErrUnauthorized)
		return
	}
	var req createCallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errwrap.Wrap(model.ErrValidation, err))
		return
	}
	if req.TableID != t.ID {
		_ = c.Error(model.ErrForbidden)
		return
	}
	wr, err := h.usecase.Create(c.Request.Context(), req.TableID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, wr)
}
