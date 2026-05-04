package profile_handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tiptop-co/backend/internal/model"
	"github.com/tiptop-co/backend/internal/model/authz"
	"github.com/tiptop-co/backend/internal/providers/http/ctxclaims"
	"github.com/tiptop-co/backend/internal/usecase/profile"
	"github.com/tiptop-co/backend/pkg/errwrap"
)

type updateProfileRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name"  binding:"required"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password"     binding:"required"`
}

type tipsResponse struct {
	Amount int `json:"amount"`
}

type ProfileHandler struct {
	usecase profile.ProfileUsecase
}

func NewProfileHandler(u profile.ProfileUsecase) *ProfileHandler {
	return &ProfileHandler{usecase: u}
}

func (h *ProfileHandler) GetMe(c *gin.Context) {
	claims := ctxclaims.GetClaims(c)
	if claims == nil {
		_ = c.Error(model.ErrUnauthorized)
		return
	}
	u, err := h.usecase.GetMe(c.Request.Context(), claims.UserID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, u)
}

func (h *ProfileHandler) Update(c *gin.Context) {
	claims := ctxclaims.GetClaims(c)
	if claims == nil {
		_ = c.Error(model.ErrUnauthorized)
		return
	}
	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errwrap.Wrap(model.ErrValidation, err))
		return
	}
	u, err := h.usecase.UpdateProfile(c.Request.Context(), claims.UserID, req.FirstName, req.LastName)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, u)
}

func (h *ProfileHandler) ChangePassword(c *gin.Context) {
	claims := ctxclaims.GetClaims(c)
	if claims == nil {
		_ = c.Error(model.ErrUnauthorized)
		return
	}
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errwrap.Wrap(model.ErrValidation, err))
		return
	}
	if err := h.usecase.ChangePassword(c.Request.Context(), claims.UserID, req.CurrentPassword, req.NewPassword); err != nil {
		_ = c.Error(err)
		return
	}
	c.Status(http.StatusOK)
}

func (h *ProfileHandler) TodayTips(c *gin.Context) {
	claims := ctxclaims.GetClaims(c)
	if claims == nil || claims.UserRole != authz.RoleWaiter {
		_ = c.Error(model.ErrForbidden)
		return
	}
	amount, err := h.usecase.GetTodayTips(c.Request.Context(), claims.UserID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, &tipsResponse{Amount: amount})
}
