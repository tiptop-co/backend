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
	"github.com/tiptop-co/backend/internal/model/authz"
	assigner_postgres "github.com/tiptop-co/backend/internal/repository/assigner/postgres"
	creds_postgres "github.com/tiptop-co/backend/internal/repository/auth/credentials/postgres"
	refresh_redis "github.com/tiptop-co/backend/internal/repository/auth/refresh-token/redis"
	menu_postgres "github.com/tiptop-co/backend/internal/repository/menu/postgres"
	order_postgres "github.com/tiptop-co/backend/internal/repository/order/postgres"
	stats_postgres "github.com/tiptop-co/backend/internal/repository/stats/postgres"
	waiter_request_postgres "github.com/tiptop-co/backend/internal/repository/waiter_request/postgres"
	table_postgres "github.com/tiptop-co/backend/internal/repository/table/postgres"
	transaction_postgres "github.com/tiptop-co/backend/internal/repository/transaction/postgres"
	user_postgres "github.com/tiptop-co/backend/internal/repository/user/postgres"
	venue_postgres "github.com/tiptop-co/backend/internal/repository/venue/postgres"
	"github.com/tiptop-co/backend/internal/usecase/auth"
	"github.com/tiptop-co/backend/internal/usecase/menu"
	orderusecase "github.com/tiptop-co/backend/internal/usecase/order"
	"github.com/tiptop-co/backend/internal/usecase/profile"
	"github.com/tiptop-co/backend/internal/usecase/table"
	tableclose "github.com/tiptop-co/backend/internal/usecase/table_close"
	txusecase "github.com/tiptop-co/backend/internal/usecase/transaction"
	statsusecase "github.com/tiptop-co/backend/internal/usecase/stats"
	userusecase "github.com/tiptop-co/backend/internal/usecase/user"
	wrusecase "github.com/tiptop-co/backend/internal/usecase/waiter_request"
	venueusecase "github.com/tiptop-co/backend/internal/usecase/venue"

	password_gen "github.com/tiptop-co/backend/internal/providers/password-gen"
	payment_stub "github.com/tiptop-co/backend/internal/providers/payment/stub"

	auth_handler "github.com/tiptop-co/backend/internal/handler/auth"
	admin_handler "github.com/tiptop-co/backend/internal/handler/admin"
	guest_handler "github.com/tiptop-co/backend/internal/handler/guest"
	manager_handler "github.com/tiptop-co/backend/internal/handler/manager"
	profile_handler "github.com/tiptop-co/backend/internal/handler/profile"
	table_handler "github.com/tiptop-co/backend/internal/handler/table"
	waiter_handler "github.com/tiptop-co/backend/internal/handler/waiter"

	"github.com/tiptop-co/backend/internal/providers/http/cookie"
	"github.com/tiptop-co/backend/internal/providers/http/middleware"
	bcrypt_hasher "github.com/tiptop-co/backend/internal/providers/password-hasher/bcrypt-hasher"
	cryptogen "github.com/tiptop-co/backend/internal/providers/tokens/crypto"
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
	tableRepo := table_postgres.NewTableRepository(pgxpool)
	menuRepo := menu_postgres.NewMenuRepository(pgxpool)
	venueRepo := venue_postgres.NewVenueRepository(pgxpool)
	userRepo := user_postgres.NewUserRepository(pgxpool)
	transactionRepo := transaction_postgres.NewTransactionRepository(pgxpool)
	orderRepo := order_postgres.NewOrderRepository(pgxpool)
	assignerRepo := assigner_postgres.NewAssignerRepository(pgxpool)
	waiterRequestRepo := waiter_request_postgres.NewRepository(pgxpool)
	statsRepo := stats_postgres.NewRepository(pgxpool)

	// PROVIDERS
	hasher := bcrypt_hasher.New(cfg.PasswordHasher.Cost)
	tokenService := jwttokens.New(cfg.AccessToken, cfg.RefreshToken)
	cookieSetter := cookie.NewCookieTokensSetter(&cfg.AuthCookie)
	tokenGenerator := cryptogen.NewGenerator()

	// USECASES
	authUsecase := auth.NewAuthService(
		credsRepo,
		refreshTokenRepo,
		hasher,
		tokenService,
	)
	tableUsecase := table.NewTableService(&cfg.TableService, tableRepo, tokenGenerator)
	menuUsecase := menu.NewMenuService(menuRepo)
	venueUsecase := venueusecase.NewVenueService(venueRepo)
	profileUsecase := profile.NewProfileService(userRepo, transactionRepo, authUsecase)

	paymentGateway := payment_stub.New()
	orderUsecase := orderusecase.NewOrderService(orderRepo, menuRepo, tableRepo, assignerRepo)
	transactionUsecase := txusecase.NewTransactionService(orderRepo, transactionRepo, tableRepo, paymentGateway)
	waiterRequestUsecase := wrusecase.NewService(waiterRequestRepo, tableRepo, assignerRepo)
	tableCloseUsecase := tableclose.NewService(tableRepo, orderRepo, tokenGenerator, cfg.TableService.SessionTokenSize)

	passwordGenerator := password_gen.New(12)
	userUsecase := userusecase.NewService(userRepo, hasher, passwordGenerator)
	statsUsecase := statsusecase.NewService(statsRepo)

	// HTTP HANDLERS
	authHandler := auth_handler.NewAuthHandler(authUsecase, cookieSetter)
	tableHandler := table_handler.NewTableHandler(cfg.TableSessionCookie, tableUsecase)
	guestMenuHandler := guest_handler.NewMenuHandler(menuUsecase)
	managerMenuHandler := manager_handler.NewMenuHandler(menuUsecase)
	managerVenueHandler := manager_handler.NewVenueHandler(venueUsecase)
	profileHandler := profile_handler.NewProfileHandler(profileUsecase)
	guestOrderHandler := guest_handler.NewOrderHandler(orderUsecase, transactionUsecase)
	guestCallHandler := guest_handler.NewCallHandler(waiterRequestUsecase)
	waiterRequestHandler := waiter_handler.NewRequestHandler(waiterRequestUsecase)
	waiterTableHandler := waiter_handler.NewTableHandler(tableUsecase, orderUsecase, waiterRequestUsecase, tableCloseUsecase)
	managerWaitersHandler := manager_handler.NewWaitersHandler(userUsecase)
	managerStatsHandler := manager_handler.NewStatsHandler(statsUsecase)
	adminManagersHandler := admin_handler.NewManagersHandler(userUsecase, venueUsecase)
	adminStatsHandler := admin_handler.NewStatsHandler(statsUsecase)

	// ROUTER
	r := gin.New()

	// middleware
	r.Use(
		gin.Recovery(),
		gin.Logger(),
		// middleware.CORSMiddleware(""),
		middleware.ParseClaims(authUsecase),
		middleware.ParseTableSession(tableUsecase),
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
	}
	users := apiV1.Group("/users")
	{
		users.PATCH("/me/password",
			middleware.RequirePermission(authz.PermUpdatePassword),
			authHandler.UpdatePassword,
		)
	}
	tables := apiV1.Group("tables")
	{
		tables.POST("/bootstrap", tableHandler.GetByQR)
		tables.GET("/:table_id/menu", middleware.RequireTableSession(), guestMenuHandler.GetMenu)
		tables.GET("/:table_id/order", middleware.RequireTableSession(), guestOrderHandler.GetOrder)
		tables.GET("/:table_id/call-status", middleware.RequireTableSession(), guestCallHandler.Status)
		tables.GET("/:table_id/requests", middleware.RequireTableSession(), guestCallHandler.List)
	}
	apiV1.GET("/dishes/:id", guestMenuHandler.GetDish)
	apiV1.POST("/orders", middleware.RequireTableSession(), guestOrderHandler.CreateOrder)
	apiV1.POST("/transactions", middleware.RequireTableSession(), guestOrderHandler.CreateTransaction)
	apiV1.POST("/requests", middleware.RequireTableSession(), guestCallHandler.Create)
	waiter := apiV1.Group("/waiter")
	{
		waiter.GET("/tables", middleware.RequirePermission(authz.PermGetWaiterTables), tableHandler.GetWaiterTables)
		waiter.GET("/tables/:table_id", middleware.RequirePermission(authz.PermGetTableByID), tableHandler.GetByID)
		waiter.POST("/tables/:table_id/close", middleware.RequirePermission(authz.PermFreeTable), tableHandler.FreeTable)

		waiter.GET("/requests", middleware.RequirePermission(authz.PermRequestListWaiter), waiterRequestHandler.List)
		waiter.POST("/requests/:id/accept", middleware.RequirePermission(authz.PermRequestAccept), waiterRequestHandler.Accept)

		waiter.GET("/tables/:table_id/detail", middleware.RequirePermission(authz.PermGetTableByID), waiterTableHandler.Detail)
		waiter.GET("/orders/completed", middleware.RequirePermission(authz.PermGetWaiterTables), waiterTableHandler.CompletedOrders)
		waiter.POST("/tables/:table_id/checkout", middleware.RequirePermission(authz.PermCloseTable), waiterTableHandler.Close)
	}
	manager := apiV1.Group("/manager")
	{
		manager.GET("/tables", middleware.RequirePermission(authz.PermGetVenueTables), tableHandler.GetVenueTables)
		manager.POST("/tables", middleware.RequirePermission(authz.PermCreateTable), tableHandler.CreateTable)
		manager.DELETE("/tables/:table_id", middleware.RequirePermission(authz.PermDeleteTable), tableHandler.DeleteTable)

		manager.GET("/menu", middleware.RequirePermission(authz.PermMenuRead), managerMenuHandler.GetMenu)
		manager.POST("/dishes", middleware.RequirePermission(authz.PermDishCreate), managerMenuHandler.CreateDish)
		manager.DELETE("/dishes/:id", middleware.RequirePermission(authz.PermDishDelete), managerMenuHandler.DeleteDish)

		manager.GET("/venue", middleware.RequirePermission(authz.PermVenueRead), managerVenueHandler.Get)
		manager.PUT("/venue", middleware.RequirePermission(authz.PermVenueUpdate), managerVenueHandler.Update)

		manager.GET("/waiters", middleware.RequirePermission(authz.PermWaiterList), managerWaitersHandler.List)
		manager.POST("/waiters", middleware.RequirePermission(authz.PermWaiterCreate), managerWaitersHandler.Create)
		manager.DELETE("/waiters/:id", middleware.RequirePermission(authz.PermWaiterDelete), managerWaitersHandler.Delete)

		manager.GET("/stats", middleware.RequirePermission(authz.PermStatsReadVenue), managerStatsHandler.Get)
	}

	admin := apiV1.Group("/admin")
	{
		admin.GET("/managers", middleware.RequirePermission(authz.PermManagerList), adminManagersHandler.List)
		admin.GET("/venues", middleware.RequirePermission(authz.PermManagerList), adminManagersHandler.ListVenues)
		admin.POST("/managers", middleware.RequirePermission(authz.PermManagerCreate), adminManagersHandler.Create)
		admin.DELETE("/managers/:id", middleware.RequirePermission(authz.PermManagerDelete), adminManagersHandler.Delete)

		admin.GET("/stats", middleware.RequirePermission(authz.PermStatsReadGlobal), adminStatsHandler.Get)
	}

	apiV1.GET("/profile/me", profileHandler.GetMe)
	apiV1.PUT("/profile", middleware.RequirePermission(authz.PermProfileUpdate), profileHandler.Update)
	apiV1.POST("/profile/password", middleware.RequirePermission(authz.PermUpdatePassword), profileHandler.ChangePassword)
	waiter.GET("/tips/today", middleware.RequirePermission(authz.PermTipsRead), profileHandler.TodayTips)

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
