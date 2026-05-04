package waiter_handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tiptop-co/backend/internal/model"
	ordermodel "github.com/tiptop-co/backend/internal/model/order"
	tablemodel "github.com/tiptop-co/backend/internal/model/table"
	wrmodel "github.com/tiptop-co/backend/internal/model/waiter_request"
	"github.com/tiptop-co/backend/internal/providers/http/ctxclaims"
	orderusecase "github.com/tiptop-co/backend/internal/usecase/order"
	tableusecase "github.com/tiptop-co/backend/internal/usecase/table"
	tableclose "github.com/tiptop-co/backend/internal/usecase/table_close"
	wrusecase "github.com/tiptop-co/backend/internal/usecase/waiter_request"
)

type tableDetailResponse struct {
	Table    *tablemodel.Table  `json:"table"`
	Order    *ordermodel.Order  `json:"order,omitempty"`
	Requests []*wrmodel.Request `json:"requests"`
}

type TableHandler struct {
	tables   tableusecase.TableUsecase
	orders   orderusecase.OrderUsecase
	requests wrusecase.Usecase
	closeUC  tableclose.Usecase
}

func NewTableHandler(t tableusecase.TableUsecase, o orderusecase.OrderUsecase, r wrusecase.Usecase, c tableclose.Usecase) *TableHandler {
	return &TableHandler{tables: t, orders: o, requests: r, closeUC: c}
}

type completedOrdersResponse struct {
	Data []*ordermodel.CompletedSummary `json:"data"`
}

func (h *TableHandler) CompletedOrders(c *gin.Context) {
	claims := ctxclaims.GetClaims(c)
	if claims == nil {
		_ = c.Error(model.ErrUnauthorized)
		return
	}
	list, err := h.orders.GetCompletedByWaiter(c.Request.Context(), claims.UserID, 50)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, &completedOrdersResponse{Data: list})
}

func (h *TableHandler) Detail(c *gin.Context) {
	claims := ctxclaims.GetClaims(c)
	if claims == nil {
		_ = c.Error(model.ErrUnauthorized)
		return
	}
	tableID := c.Param("table_id")
	if tableID == "" {
		_ = c.Error(model.ErrValidation)
		return
	}

	t, err := h.tables.GetByID(c.Request.Context(), tableID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	if t.WaiterID != nil && *t.WaiterID != claims.UserID {
		_ = c.Error(model.ErrForbidden)
		return
	}

	o, err := h.orders.GetByTable(c.Request.Context(), tableID)
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		_ = c.Error(err)
		return
	}

	reqs, err := h.requests.GetByTable(c.Request.Context(), tableID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, &tableDetailResponse{Table: t, Order: o, Requests: reqs})
}

func (h *TableHandler) Close(c *gin.Context) {
	claims := ctxclaims.GetClaims(c)
	if claims == nil {
		_ = c.Error(model.ErrUnauthorized)
		return
	}
	tableID := c.Param("table_id")
	if tableID == "" {
		_ = c.Error(model.ErrValidation)
		return
	}

	t, err := h.tables.GetByID(c.Request.Context(), tableID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	if t.WaiterID != nil && *t.WaiterID != claims.UserID {
		_ = c.Error(model.ErrForbidden)
		return
	}

	closed, err := h.closeUC.Close(c.Request.Context(), tableID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, closed)
}
