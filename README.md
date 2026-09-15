# Ajora

> Traditional Ajo on the surface. Web3 underneath.

Digital community savings (Ajo / Esusu) platform for Nigeria.

## Status

**Phase 1 — Foundation.** No application features yet. This repo currently
contains: repo skeleton, Go API entrypoint with `/health` and `/ready`,
the canonical `money` package, Docker Compose for Postgres + Redis,
Expo + TypeScript mobile app scaffold, and docs.

## Layout

| Path         | Purpose                                       |
|--------------|-----------------------------------------------|
| `api/`       | Go backend (Gin, pgx, Redis, WebSockets)      |
| `mobile/`    | React Native (Expo + TypeScript)              |
| `contracts/` | Solidity smart contracts (Foundry) — Phase 7  |
| `infra/`     | Docker Compose, Nginx, Terraform, CI          |
| `docs/`      | Product, UX, API, DB, security, runbooks      |
| `scripts/`   | One-off dev utilities                         |

## Getting started

```bash
cp .env.example .env            # then fill in secrets
make up                         # Postgres + Redis
make api-test                   # Go tests
make api-run                    # API on :8080
make mobile-install && make mobile-start


cat > api/internal/config/config.go << 'EOF'
// Package config loads and validates Ajora runtime configuration.
package config

import (
    "fmt"
    "os"
    "strconv"
    "time"
)

type Config struct {
    Env  string
    Name string

    Port int

    DatabaseURL string
    RedisURL    string

    JWTAccessSecret  string
    JWTRefreshSecret string
    JWTAccessTTL     time.Duration
    JWTRefreshTTL    time.Duration

    OTPTTL         time.Duration
    OTPMaxAttempts int

    LogLevel string
}

func Load() (*Config, error) {
    c := &Config{
        Env:              getEnv("APP_ENV", "development"),
        Name:             getEnv("APP_NAME", "ajora"),
        Port:             getEnvInt("API_PORT", 8080),
        DatabaseURL:      getEnv("DATABASE_URL", ""),
        RedisURL:         getEnv("REDIS_URL", "redis://localhost:6379/0"),
        JWTAccessSecret:  getEnv("JWT_ACCESS_SECRET", ""),
        JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", ""),
        JWTAccessTTL:     getEnvDuration("JWT_ACCESS_TTL", 15*time.Minute),
        JWTRefreshTTL:    getEnvDuration("JWT_REFRESH_TTL", 720*time.Hour),
        OTPTTL:           getEnvDuration("OTP_TTL", 5*time.Minute),
        OTPMaxAttempts:   getEnvInt("OTP_MAX_ATTEMPTS", 5),
        LogLevel:         getEnv("LOG_LEVEL", "debug"),
    }

    if c.Env == "production" {
        if c.JWTAccessSecret == "" || len(c.JWTAccessSecret) < 32 {
            return nil, fmt.Errorf("JWT_ACCESS_SECRET must be >= 32 chars in production")
        }
        if c.JWTRefreshSecret == "" || len(c.JWTRefreshSecret) < 32 {
            return nil, fmt.Errorf("JWT_REFRESH_SECRET must be >= 32 chars in production")
        }
        if c.DatabaseURL == "" {
            return nil, fmt.Errorf("DATABASE_URL is required in production")
        }
    }

    return c, nil
}

func (c *Config) IsProd() bool { return c.Env == "production" }

func getEnv(k, def string) string {
    if v := os.Getenv(k); v != "" {
        return v
    }
    return def
}

func getEnvInt(k string, def int) int {
    if v := os.Getenv(k); v != "" {
        if n, err := strconv.Atoi(v); err == nil {
            return n
        }
    }
    return def
}

func getEnvDuration(k string, def time.Duration) time.Duration {
    if v := os.Getenv(k); v != "" {
        if d, err := time.ParseDuration(v); err == nil {
            return d
        }
    }
    return def
}
