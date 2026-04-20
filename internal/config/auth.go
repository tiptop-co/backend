package config

import "time"

type AccessTokenConfig struct {
	TTL       time.Duration `yaml:"access_token_ttl" env:"ACCESS_TOKEN_TTL"`
	Jitter    time.Duration `yaml:"access_tolen_jitter" env:"ACCESS_TOKEN_JITTER"`
	SecretKey []byte        `yaml:"secret_key" env:"SECRET_KEY"`
}

type RefreshTokenConfig struct {
	Size   int           `env:"REFRESH_TOKEN_SIZE"`
	TTL    time.Duration `env:"REFRESH_TOKEN_TTL"`
	Jitter time.Duration `env:"REFRESH_TOKEN_JITTER"`
}

type PasswordHasherConfig struct {
	Cost int `env:"PASSWORD_HASHER_COST"`
}

type AuthCookieConfig struct {
	Domain     string        `env:"AUTH_COOKIE_DOMAIN"`
	Path       string        `env:"AUTH_COOKIE_PATH"`
	Secure     bool          `env:"AUTH_COOKIE_SECURE"`
	HttpOnly   bool          `env:"AUTH_COOKIE_HTTP_ONLY"`
	RefreshTTL time.Duration `env:"AUTH_COOKIE_REFRESH_TTL"`
	AccessTTL  time.Duration `env:"AUTH_COOKIE_ACCESS_TTL"`
}
