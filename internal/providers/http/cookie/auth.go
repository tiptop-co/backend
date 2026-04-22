package cookie

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tiptop-co/backend/internal/config"
	"github.com/tiptop-co/backend/internal/model/auth"
)

const (
	RefreshCookieName = "refresh_token"
	AccessCookieName  = "access_token"
)

type CookieTokensSetter struct {
	cookieConfig *config.AuthCookieConfig
}

func NewCookieTokensSetter(cookieConfig *config.AuthCookieConfig) *CookieTokensSetter {
	return &CookieTokensSetter{cookieConfig: cookieConfig}
}

func (c *CookieTokensSetter) SetAll(ctx *gin.Context, tokens *auth.AuthTokens) {
	c.setToken(ctx, RefreshCookieName, tokens.Refresh, c.cookieConfig.RefreshTTL)
	c.setToken(ctx, AccessCookieName, tokens.Access, c.cookieConfig.AccessTTL)
}

func (c *CookieTokensSetter) SetRefresh(ctx *gin.Context, refreshToken string) {
	c.setToken(ctx, RefreshCookieName, refreshToken, c.cookieConfig.RefreshTTL)
}

func (c *CookieTokensSetter) SetAccess(ctx *gin.Context, accessToken string) {
	c.setToken(ctx, AccessCookieName, accessToken, c.cookieConfig.AccessTTL)
}

func (c *CookieTokensSetter) ResetRefresh(ctx *gin.Context) {
	c.setToken(ctx, RefreshCookieName, "", 0)
}

func (c *CookieTokensSetter) ResetAccess(ctx *gin.Context) {
	c.setToken(ctx, AccessCookieName, "", 0)
}

func (c *CookieTokensSetter) ResetAll(ctx *gin.Context) {
	c.ResetAccess(ctx)
	c.ResetRefresh(ctx)
}

func (c *CookieTokensSetter) setToken(ctx *gin.Context, tokenName string, token string, ttl time.Duration) {
	ctx.SetCookie(
		tokenName,
		token,
		int(ttl.Seconds()),
		c.cookieConfig.Path,     // path
		c.cookieConfig.Domain,   // domain ("" = current)
		c.cookieConfig.Secure,   // secure (set to false if testing locally w/o HTTPS)
		c.cookieConfig.HttpOnly, // httpOnly
	)
}
