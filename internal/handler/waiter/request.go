package waiter_handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tiptop-co/backend/internal/model"
	wrmodel "github.com/tiptop-co/backend/internal/model/waiter_request"
	"github.com/tiptop-co/backend/internal/providers/http/ctxclaims"
	wrusecase "github.com/tiptop-co/backend/internal/usecase/waiter_request"
)

type requestsListResponse struct {
	Data []*wrmodel.Request `json:"data"`
}

type RequestHandler struct {
	usecase wrusecase.Usecase
}

func NewRequestHandler(u wrusecase.Usecase) *RequestHandler {
	return &RequestHandler{usecase: u}
}

func (h *RequestHandler) List(c *gin.Context) {
	claims := ctxclaims.GetClaims(c)
	if claims == nil {
		_ = c.Error(model.ErrUnauthorized)
		return
	}
	reqs, err := h.usecase.GetByWaiter(c.Request.Context(), claims.UserID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, &requestsListResponse{Data: reqs})
}

func (h *RequestHandler) Accept(c *gin.Context) {
	claims := ctxclaims.GetClaims(c)
	if claims == nil {
		_ = c.Error(model.ErrUnauthorized)
		return
	}
	id := c.Param("id")
	if id == "" {
		_ = c.Error(model.ErrValidation)
		return
	}
	wr, err := h.usecase.Accept(c.Request.Context(), id, claims.UserID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, wr)
}
