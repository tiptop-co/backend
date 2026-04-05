package configure

import (
	"github.com/go-redis/redis"

	"github.com/tiptop-co/backend/internal/config"
)

const defaultDB = 0

func MustInitRedis(cfg config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: cfg.Addr,
		DB:   defaultDB,
	})
}
