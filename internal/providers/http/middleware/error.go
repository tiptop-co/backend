package middleware

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tiptop-co/backend/internal/model"
	jwttokens "github.com/tiptop-co/backend/internal/providers/tokens/jwt"
	"github.com/tiptop-co/backend/internal/usecase/auth"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func ErrorHandler(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		logger.Error("request failed",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"client_ip", c.ClientIP(),
			"error", err,
		)

		status, code := mapError(err)

		c.AbortWithStatusJSON(status, ErrorResponse{
			Error: code,
		})
	}
}

func mapError(err error) (int, string) {
	switch {

	// AUTH
	case errors.Is(err, auth.ErrInvalidCredentials),
		errors.Is(err, auth.ErrWrongOldPassword):
		return http.StatusUnauthorized, "invalid_credentials"

	case errors.Is(err, jwttokens.ErrExpired):
		return http.StatusUnauthorized, "token_expired"

	case errors.Is(err, jwttokens.ErrInvalidAccessToken),
		errors.Is(err, jwttokens.ErrInvalidClaims),
		errors.Is(err, jwttokens.ErrInvalidSigningMethod):
		return http.StatusUnauthorized, "invalid_token"

	// COMMON
	case errors.Is(err, model.ErrNotFound):
		return http.StatusNotFound, "not_found"

	case errors.Is(err, model.ErrValidation):
		return http.StatusBadRequest, "bad_request"

	case errors.Is(err, model.ErrUnauthorized):
		return http.StatusBadRequest, "unauthorized"

	case errors.Is(err, model.ErrForbidden):
		return http.StatusForbidden, "forbidden"

	default:
		return http.StatusInternalServerError, "internal_error"
	}
}
