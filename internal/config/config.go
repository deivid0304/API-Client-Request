package config

import (
	"errors"
	"os"
	"time"
)

type Config struct {
	Port            string
	DatabaseURL     string
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	SessionTTL      time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		AccessTokenTTL:  mustDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL: mustDuration("REFRESH_TOKEN_TTL", 7*24*time.Hour),
		SessionTTL:      mustDuration("SESSION_TTL", 30*24*time.Hour),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL nao configurada")
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, errors.New("JWT_SECRET deve ter pelo menos 32 caracteres")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func mustDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
