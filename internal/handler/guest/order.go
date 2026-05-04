package guest_handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tiptop-co/backend/internal/model"
	ordermodel "github.com/tiptop-co/backend/internal/model/order"
	"github.com/tiptop-co/backend/internal/providers/http/middleware"
	orderusecase "github.com/tiptop-co/backend/internal/usecase/order"
	txusecase "github.com/tiptop-co/backend/internal/usecase/transaction"
	"github.com/tiptop-co/backend/pkg/errwrap"
)

type createOrderItem struct {
	DishID   string `json:"dish_id"  binding:"required"`
	Quantity int    `json:"quantity" binding:"required"`
}

type createOrderRequest struct {
	TableID string            `json:"table_id" binding:"required"`
	Items   []createOrderItem `json:"items"    binding:"required,min=1"`
	Wishes  *string           `json:"wishes"`
}

type createTransactionRequest struct {
	OrderID    string   `json:"order_id"    binding:"required"`
	ItemIDs    []string `json:"item_ids"    binding:"required,min=1"`
	TipsAmount int      `json:"tips_amount"`
}

type OrderHandler struct {
	orders orderusecase.OrderUsecase
	tx     txusecase.TransactionUsecase
}

func NewOrderHandler(orders orderusecase.OrderUsecase, tx txusecase.TransactionUsecase) *OrderHandler {
	return &OrderHandler{orders: orders, tx: tx}
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
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

	o, err := h.orders.GetByTable(c.Request.Context(), tableID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, o)
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	t := middleware.GetTable(c)
	if t == nil {
		_ = c.Error(model.ErrUnauthorized)
		return
	}
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errwrap.Wrap(model.ErrValidation, err))
		return
	}
	if req.TableID != t.ID {
		_ = c.Error(model.ErrForbidden)
		return
	}

	items := make([]ordermodel.CreateItemInput, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, ordermodel.CreateItemInput{DishID: it.DishID, Quantity: it.Quantity})
	}

	o, err := h.orders.Create(c.Request.Context(), req.TableID, items, req.Wishes)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, o)
}

func (h *OrderHandler) CreateTransaction(c *gin.Context) {
	t := middleware.GetTable(c)
	if t == nil {
		_ = c.Error(model.ErrUnauthorized)
		return
	}
	var req createTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errwrap.Wrap(model.ErrValidation, err))
		return
	}

	txn, err := h.tx.Create(c.Request.Context(), req.OrderID, req.ItemIDs, req.TipsAmount)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, txn)
}
