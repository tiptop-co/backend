package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	apiMiddleware "github.com/go-openapi/runtime/middleware"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"

	"github.com/tiptop-co/backend/internal/config"
	"github.com/tiptop-co/backend/internal/configure"
	mealsHandler "github.com/tiptop-co/backend/internal/handler/v1_meals"
	"github.com/tiptop-co/backend/internal/model"
	authMiddleware "github.com/tiptop-co/backend/internal/pkg/middleware/auth"
	paymentRepo "github.com/tiptop-co/backend/internal/repository/payment"
	payOrderUc "github.com/tiptop-co/backend/internal/usecase/order/pay"
)

const timeout = 3 * time.Second

func main() {
	cfg := config.MustLoad()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	db := configure.MustInitPostgres(ctx, cfg.Postgres)

	redis := configure.MustInitRedis(cfg.Redis)

	reg := prometheus.NewRegistry()

	// <! Repositories
	paymentRepository := paymentRepo.New(db)
	// !>

	// <! Usecases
	payOrderUsecase := payOrderUc.New(paymentRepository)
	// !>

	// <! Router
	r := mux.NewRouter()
	// !!>

	specData, err := os.ReadFile("api/schema.yml")
	if err != nil {
		panic(err)
	}

	r.Handle("/api/schema.yml", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write(specData)
	}))

	opts := apiMiddleware.SwaggerUIOpts{
		Path:    "/docs",
		SpecURL: "/api/schema.yml",
		Title:   "API Documentation",
	}

	r.Handle("/docs/", apiMiddleware.SwaggerUI(opts, nil))

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL)
		http.Error(w, `Not found!`, 404)
	})

	r.Handle("/public/metrics", promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{
			Registry: reg,
		},
	))

	r.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	// !>

	// <! Permissions
	permsMiddleware := permissionMiddleware.New()

	for path, perms := range model.Resources {
		permsMiddleware.Register(path, perms)
	}
	// !>

	// <! Middleware
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Content-Length"},
		AllowCredentials: true,
	})
	r.Use(authMiddleware.New())
	// r.Use(permsMiddleware.New())
	// !>

	// <! Handlers
	authHandler := mealsHandler.New()
	// !>

	// <! Server
	srv := server.NewServer(corsMiddleware.Handler(srvRouter), cfg.Server)
	// !>

	// <! Run
	go func() {
		log.Info(fmt.Sprintf("server is running on port %s...", cfg.Server.Port))
		if err := srv.Run(); err != nil {
			log.Error(">>> ERROR: HTTP server ListenAndServe error: " + err.Error())
		}
	}()
	// !>

	// <! Graceful shutdown
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-exit

	ctx, shutdown := context.WithTimeout(context.Background(), timeout)
	defer shutdown()

	log.Info("shutting down...")
	if err := srv.Stop(ctx); err != nil {
		log.Error(fmt.Sprintf("HTTP server shutdown error: %v", err))
	}
	// !>
}
