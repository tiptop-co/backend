package config

import (
	"time"
)

type TableServiceConfig struct {
	QRTokenSize      int `yaml:"qr_token_size" env:"TABLE_QR_TOKEN_SIZE"`
	SessionTokenSize int `yaml:"session_token_size" env:"TABLE_SESSION_TOKEN_SIZE"`
}

type TableSessionCookieConfig struct {
	Domain   string        `env:"TABLE_SESSION_COOKIE_DOMAIN"`
	Path     string        `env:"TABLE_SESSION_COOKIE_PATH"`
	Secure   bool          `env:"TABLE_SESSION_COOKIE_SECURE"`
	HttpOnly bool          `env:"TABLE_SESSION_COOKIE_HTTP_ONLY"`
	TTL      time.Duration `env:"TABLE_SESSION_COOKIE_TTL"`
}
