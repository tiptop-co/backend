package admin_handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tiptop-co/backend/internal/model"
	usermodel "github.com/tiptop-co/backend/internal/model/user"
	venuemodel "github.com/tiptop-co/backend/internal/model/venue"
	userusecase "github.com/tiptop-co/backend/internal/usecase/user"
	venueusecase "github.com/tiptop-co/backend/internal/usecase/venue"
	"github.com/tiptop-co/backend/pkg/errwrap"
)

type createManagerRequest struct {
	FirstName string  `json:"first_name" binding:"required"`
	LastName  string  `json:"last_name"  binding:"required"`
	Phone     string  `json:"phone"      binding:"required"`
	VenueID   *string `json:"venue_id"`
}

type managersListResponse struct {
	Data []*usermodel.User `json:"data"`
}

type venuesListResponse struct {
	Data []*venuemodel.Venue `json:"data"`
}

type ManagersHandler struct {
	usecase userusecase.ManagerUsecase
	venues  venueusecase.VenueUsecase
}

func NewManagersHandler(u userusecase.ManagerUsecase, v venueusecase.VenueUsecase) *ManagersHandler {
	return &ManagersHandler{usecase: u, venues: v}
}

func (h *ManagersHandler) List(c *gin.Context) {
	users, err := h.usecase.GetManagers(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, &managersListResponse{Data: users})
}

func (h *ManagersHandler) ListVenues(c *gin.Context) {
	venues, err := h.venues.GetAll(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, &venuesListResponse{Data: venues})
}

func (h *ManagersHandler) Create(c *gin.Context) {
	var req createManagerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errwrap.Wrap(model.ErrValidation, err))
		return
	}
	if req.VenueID != nil && *req.VenueID == "" {
		req.VenueID = nil
	}
	res, err := h.usecase.CreateManager(c.Request.Context(), req.VenueID, req.FirstName, req.LastName, req.Phone)
	if err != nil {
		_ = c.Error(err)
		return
	}
	if req.VenueID != nil {
		if err := h.venues.AssignManager(c.Request.Context(), *req.VenueID, res.User.ID); err != nil {
			_ = c.Error(err)
			return
		}
	}
	c.JSON(http.StatusCreated, res)
}

func (h *ManagersHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		_ = c.Error(model.ErrValidation)
		return
	}
	if err := h.usecase.DeleteManager(c.Request.Context(), id); err != nil {
		_ = c.Error(err)
		return
	}
	c.Status(http.StatusOK)
}
