package config

import (
    "os"
    "time"
)

type Config struct {
    Env         string
    Port        string
    DatabaseURL string
    RedisURL    string
    JWTSecret   string
    AccessTTL   time.Duration
    RefreshTTL  time.Duration
}

func Load() *Config {
    return &Config{
        Env:         getEnv("ENV", "development"),
        Port:        getEnv("PORT", "8080"),
        DatabaseURL: getEnv("DATABASE_URL", "host=postgres user=ajora password=ajora_secret dbname=ajora port=5432 sslmode=disable"),
        RedisURL:    getEnv("REDIS_URL", "redis://redis:6379/0"),
        JWTSecret:   getEnv("JWT_SECRET", "change_me_in_production_please"),
        AccessTTL:   parseDuration(getEnv("ACCESS_TTL", "15m")),
        RefreshTTL:  parseDuration(getEnv("REFRESH_TTL", "168h")),
    }
}

func getEnv(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}

func parseDuration(s string) time.Duration {
    if d, err := time.ParseDuration(s); err == nil {
        return d
    }
    return 15 * time.Minute
}

func (c *Config) IsProd() bool { return c.Env == "production" }
