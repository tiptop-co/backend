package infra

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/tiptop-co/backend/internal/config"
)

func MustInitPostgres(ctx context.Context, cfg config.PostgresConfig) *pgxpool.Pool {
	hostPort := net.JoinHostPort(cfg.Host, cfg.Port)

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=%s&timezone=Europe/Moscow",
		cfg.User,
		cfg.Password,
		hostPort,
		cfg.Database,
		cfg.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("unable to parse config: %v", err)
	}
	poolConfig.MaxConns = int32(cfg.MaxOpenConns)

	ctxWithTimeout, cancel := context.WithTimeout(ctx, cfg.StartTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctxWithTimeout, poolConfig)
	if err != nil {
		log.Fatalf("unable to create pgx pool: %v", err)
	}

	if err = pool.Ping(ctxWithTimeout); err != nil {
		log.Fatalf("unable to ping database: %v", err)
	}

	return pool
}
