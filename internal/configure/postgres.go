package configure

import (
	"context"
	"fmt"
	"net"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/tiptop-co/backend/internal/config"
)

func MustInitPostgres(ctx context.Context, cfg config.PostgresConfig) *sqlx.DB {
	hostPort := net.JoinHostPort(cfg.Host, cfg.Port)

	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s&timezone=Europe/Moscow",
		cfg.User,
		cfg.Password,
		hostPort,
		cfg.Database,
		cfg.Sslmode,
	)

	db, err := sqlx.ConnectContext(ctx, "pgx", dsn)
	if err != nil {
		panic(fmt.Errorf("failed to connect to postgres: %w", err))
	}

	err = db.Ping()
	if err != nil {
		panic(fmt.Errorf("failed to ping database: %w", err))
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)

	return db
}
