package table_handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tiptop-co/backend/internal/config"
	"github.com/tiptop-co/backend/internal/model"
	"github.com/tiptop-co/backend/internal/model/table"
	"github.com/tiptop-co/backend/internal/providers/http/ctxclaims"
	usecase "github.com/tiptop-co/backend/internal/usecase/table"
	"github.com/tiptop-co/backend/pkg/errwrap"
)

var (
	ErrTableIDRequired = errors.New("table_id required")
	ErrQRRequired      = errors.New("qr_token required")
)

type TableHandler struct {
	usecase usecase.TableUsecase
	cfg     config.TableSessionCookieConfig
}

func NewTableHandler(cfg config.TableSessionCookieConfig, usecase usecase.TableUsecase) *TableHandler {
	return &TableHandler{
		usecase: usecase,
		cfg:     cfg,
	}
}

func (h *TableHandler) CreateTable(c *gin.Context) {
	var request createTableRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		_ = c.Error(errwrap.Wrap(model.ErrValidation, err))
		return
	}
	if request.Number <= 0 {
		_ = c.Error(model.ErrValidation)
		return
	}

	claims := ctxclaims.GetClaims(c)

	table, err := h.usecase.Create(c.Request.Context(), claims.VenueID, request.Number)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, table)
}

func (h *TableHandler) GetByQR(c *gin.Context) {
	var request tableByQRRequest
	if err := c.ShouldBindJSON(&request); err != nil || request.QRToken == "" {
		_ = c.Error(errwrap.Wrap(model.ErrValidation, ErrQRRequired))
		return
	}

	tables, err := h.usecase.GetByFilters(c.Request.Context(), &table.TableFilters{
		QRToken: &request.QRToken,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}
	if len(tables) == 0 {
		_ = c.Error(model.ErrNotFound)
		return
	}

	t := tables[0]
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"table_session",
		t.SessionToken,
		int(h.cfg.TTL.Seconds()),
		h.cfg.Path,
		h.cfg.Domain,
		h.cfg.Secure,
		h.cfg.HttpOnly,
	)
	c.JSON(http.StatusOK, t)
}

func (h *TableHandler) GetByID(c *gin.Context) {
	tableID := c.Param("table_id")
	if tableID == "" {
		_ = c.Error(errwrap.Wrap(model.ErrValidation, ErrTableIDRequired))
		return
	}

	table, err := h.usecase.GetByID(c.Request.Context(), tableID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, table)
}

func (h *TableHandler) GetWaiterTables(c *gin.Context) {
	claims := ctxclaims.GetClaims(c)

	tables, err := h.usecase.GetByFilters(c.Request.Context(), &table.TableFilters{
		WaiterID: &claims.UserID,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, &tablesResponse{
		Tables: tables,
	})
}

func (h *TableHandler) GetVenueTables(c *gin.Context) {
	claims := ctxclaims.GetClaims(c)

	tables, err := h.usecase.GetByFilters(c.Request.Context(), &table.TableFilters{
		VenueID: &claims.VenueID,
	})
	if err != nil {
		_ = c.Error(err)
	}

	c.JSON(http.StatusOK, &tablesResponse{
		Tables: tables,
	})
}

func (h *TableHandler) FreeTable(c *gin.Context) {
	tableID := c.Param("table_id")
	if tableID == "" {
		_ = c.Error(errwrap.Wrap(model.ErrValidation, ErrTableIDRequired))
		return
	}

	err := h.usecase.UpdateStatus(c.Request.Context(), tableID, table.StatusFree)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *TableHandler) DeleteTable(c *gin.Context) {
	tableID := c.Param("table_id")
	if tableID == "" {
		_ = c.Error(errwrap.Wrap(model.ErrValidation, ErrTableIDRequired))
		return
	}

	err := h.usecase.Delete(c.Request.Context(), tableID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusOK)
}
