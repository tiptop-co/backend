package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tiptop-co/backend/internal/model"
	"github.com/tiptop-co/backend/internal/model/auth"
	"github.com/tiptop-co/backend/internal/providers/http/cookie"
	"github.com/tiptop-co/backend/internal/providers/http/ctxclaims"
	usecase "github.com/tiptop-co/backend/internal/usecase/auth"
	"github.com/tiptop-co/backend/pkg/errwrap"
)

type AuthHandler struct {
	usecase            usecase.AuthUsecase
	cookieTokensSetter *cookie.CookieTokensSetter
}

func NewAuthHandler(usecase usecase.AuthUsecase,
	cookieTokensSetter *cookie.CookieTokensSetter) *AuthHandler {

	return &AuthHandler{
		usecase:            usecase,
		cookieTokensSetter: cookieTokensSetter,
	}
}

func (ac *AuthHandler) Login(c *gin.Context) {
	var creds auth.UserCredentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.Error(errwrap.Wrap(model.ErrValidation, err))
		return
	}

	tokens, err := ac.usecase.Login(c.Request.Context(), &creds)
	if err != nil {
		c.Error(err)
		return
	}

	ac.cookieTokensSetter.SetAll(c, tokens)
	c.Status(http.StatusOK)
}

func (ac *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie(cookie.RefreshCookieName)
	if err != nil {
		c.Error(errwrap.Wrap(model.ErrValidation, err))
		return
	}

	tokens, err := ac.usecase.RefreshTokens(c.Request.Context(), refreshToken)
	if err != nil {
		c.Error(err)
		return
	}

	ac.cookieTokensSetter.SetAll(c, tokens)
	c.Status(http.StatusOK)
}

func (ac *AuthHandler) Logout(c *gin.Context) {
	refreshToken, err := c.Cookie(cookie.RefreshCookieName)
	if err == nil {
		if err := ac.usecase.Logout(c.Request.Context(), refreshToken); err != nil {
			c.Error(err)
			return
		}
	}

	ac.cookieTokensSetter.ResetAll(c)
	c.Status(http.StatusOK)
}

func (ac *AuthHandler) UpdatePassword(c *gin.Context) {
	claims := ctxclaims.GetClaims(c)
	if claims == nil {
		c.Error(errwrap.WrapMsg("failed to get claims", model.ErrUnauthorized))
		return
	}

	var passwords auth.UserPasswords
	if err := c.ShouldBindJSON(&passwords); err != nil {
		c.Error(errwrap.Wrap(model.ErrValidation, err))
		return
	}

	if err := ac.usecase.UpdatePassword(c.Request.Context(), claims.UserID, &passwords); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusOK)
}
