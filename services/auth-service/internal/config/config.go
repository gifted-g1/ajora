
package config

import (
    "os"
    "strconv"
    "time"
)

type Config struct {
    Port            string
    DatabaseURL     string
    RedisAddr       string
    RedisPassword   string
    RedisDB         int
    JWTSecret       string
    JWTExpiry       time.Duration
    RefreshExpiry   time.Duration
    MFAIssuer       string
    EmailFrom       string
    EmailHost       string
    EmailPort       int
    EmailUser       string
    EmailPassword   string
}

func Load() *Config {
    return &Config{
        Port:           getEnv("PORT", "8081"),
        DatabaseURL:    getEnv("DATABASE_URL", "postgres://ajora_admin:password@localhost:5432/ajora?sslmode=disable"),
        RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
        RedisPassword:  getEnv("REDIS_PASSWORD", ""),
        RedisDB:        getEnvInt("REDIS_DB", 0),
        JWTSecret:      getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
        JWTExpiry:      getEnvDuration("JWT_EXPIRY", 1*time.Hour),
        RefreshExpiry:  getEnvDuration("REFRESH_EXPIRY", 168*time.Hour),
        MFAIssuer:      getEnv("MFA_ISSUER", "Ajora"),
        EmailFrom:      getEnv("EMAIL_FROM", "noreply@ajora.com"),
        EmailHost:      getEnv("EMAIL_HOST", "smtp.gmail.com"),
        EmailPort:      getEnvInt("EMAIL_PORT", 587),
        EmailUser:      getEnv("EMAIL_USER", ""),
        EmailPassword:  getEnv("EMAIL_PASSWORD", ""),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intVal, err := strconv.Atoi(value); err == nil {
            return intVal
        }
    }
    return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
    if value := os.Getenv(key); value != "" {
        if duration, err := time.ParseDuration(value); err == nil {
            return duration
        }
    }
    return defaultValue
}

