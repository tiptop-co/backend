package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	HTTP     HTTPConfig     `yaml:"http"`
	Postgres PostgresConfig `yaml:"postgres"`
	Redis    RedisConfig    `yaml:"redis"`

	// AUTH
	AccessToken    AccessTokenConfig    `yaml:"access_token"`
	RefreshToken   RefreshTokenConfig   `yaml:"refresh_token"`
	PasswordHasher PasswordHasherConfig `yaml:"password_hasher"`
	AuthCookie     AuthCookieConfig     `yaml:"auth_cookie"`
}

type HTTPConfig struct {
	Host string `yaml:"host" env:"HTTP_HOST"`
	Port string `yaml:"port" env:"HTTP_PORT"`
}

type PostgresConfig struct {
	Host         string        `yaml:"host" env:"POSTGRES_HOST"`
	Port         string        `yaml:"port" env:"POSTGRES_PORT"`
	User         string        `yaml:"user" env:"POSTGRES_USER"`
	Password     string        `yaml:"password" env:"POSTGRES_PASSWORD"`
	Database     string        `yaml:"database" env:"POSTGRES_DATABASE"`
	SSLMode      string        `yaml:"sslmode" env:"POSTGRES_SSLMODE"`
	MaxOpenConns int           `yaml:"max_open_conns" env:"POSTGRES_MAX_OPEN_CONNS"`
	StartTimeout time.Duration `env:"POSTGRES_START_TIMEOUT"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr" env:"REDIS_ADDR"`
	Password string `yaml:"password" env:"REDIS_PASSWORD"`
	DB       int    `yaml:"db" env:"REDIS_DB"`
}

func MustLoad() *Config {
	var cfg Config

	err := env.Parse(&cfg)
	if err != nil {
		panic("error while reading config: " + err.Error())
	}

	return &cfg
}
