package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/tiptop-co/backend/internal/config"
	"github.com/tiptop-co/backend/internal/infra"
	creds_postgres "github.com/tiptop-co/backend/internal/repository/auth/credentials/postgres"
	refresh_redis "github.com/tiptop-co/backend/internal/repository/auth/refresh-token/redis"
	"github.com/tiptop-co/backend/internal/usecase/auth"

	authHandler "github.com/tiptop-co/backend/internal/handler/auth"

	"github.com/tiptop-co/backend/internal/providers/http/cookie"
	"github.com/tiptop-co/backend/internal/providers/http/middleware"
	bcrypt_hasher "github.com/tiptop-co/backend/internal/providers/password-hasher/bcrypt-hasher"
	jwttokens "github.com/tiptop-co/backend/internal/providers/tokens/jwt"
)

const shutdownTimeout = 5 * time.Second

func main() {
	cfg := config.MustLoad()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	pgxpool := infra.MustInitPostgres(ctx, cfg.Postgres)
	defer pgxpool.Close()

	redisClient, err := infra.NewRedisClient(cfg.Redis)
	if err != nil {
		slog.Error("failed to connect to redis", "error", err)
		return
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.Error("failed to close redis connection", "error", err)
		}
	}()

	// REPOSITORIES
	credsRepo := creds_postgres.NewRepository(pgxpool)
	refreshTokenRepo := refresh_redis.NewRefreshTokenRepository(redisClient, &cfg.RefreshToken)

	// PROVIDERS
	hasher := bcrypt_hasher.New(cfg.PasswordHasher.Cost)
	tokenService := jwttokens.New(cfg.AccessToken, cfg.RefreshToken)
	cookieSetter := cookie.NewCookieTokensSetter(&cfg.AuthCookie)

	// USECASES
	authUsecase := auth.NewAuthService(
		credsRepo,
		refreshTokenRepo,
		hasher,
		tokenService,
	)

	// HTTP HANDLERS
	authHandler := authHandler.NewAuthHandler(authUsecase, cookieSetter)

	// ROUTER
	r := gin.New()

	// middleware
	r.Use(
		gin.Recovery(),
		gin.Logger(),
		// middleware.CORSMiddleware(""),
		middleware.ErrorHandler(logger),
	)

	// health
	r.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// ROUTES
	apiV1 := r.Group("/api/v1")
	auth := apiV1.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
		auth.POST("/logout", authHandler.Logout)
		auth.POST("/password", middleware.ClaimsRequired(authUsecase),
			authHandler.UpdatePassword)
	}

	// SERVER
	srv := &http.Server{
		Addr:    net.JoinHostPort(cfg.HTTP.Host, cfg.HTTP.Port),
		Handler: r,
	}

	go func() {
		logger.Info("server started", "port", cfg.HTTP.Port)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
		}
	}()

	// GRACEFUL SHUTDOWN
	<-ctx.Done()
	logger.Info("shutting down...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		logger.Error("shutdown failed", "error", err)
	}

	logger.Info("server stopped")
}
