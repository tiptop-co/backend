package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Service  ServiceConfig  `yaml:"service"`
	Postgres PostgresConfig `yaml:"postgres"`
	Redis    RedisConfig    `yaml:"redis"`
}

type ServiceConfig struct {
	SessionTTL time.Duration `yaml:"session_ttl" env:"SESSION_TTL"`
}

type PostgresConfig struct {
	Host         string `yaml:"host" env:"POSTGRES_HOST"`
	Port         string `yaml:"port" env:"POSTGRES_PORT"`
	User         string `yaml:"user" env:"POSTGRES_USER"`
	Password     string `yaml:"password" env:"POSTGRES_PASSWORD"`
	Database     string `yaml:"database" env:"POSTGRES_DATABASE"`
	Sslmode      string `yaml:"sslmode" env:"POSTGRES_SSLMODE"`
	MaxOpenConns int    `yaml:"max_open_conns" env:"POSTGRES_MAX_OPEN_CONNS"`
}

type RedisConfig struct {
	Addr string `yaml:"addr" env:"REDIS_ADDR"`
}

func MustLoad() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file reading error:", err.Error())
	}

	var cfg Config

	err = cleanenv.ReadEnv(&cfg)
	if err != nil {
		panic("error while reading config: " + err.Error())
	}

	return &cfg
}
