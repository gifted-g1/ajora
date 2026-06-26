
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/ajora/auth-service/internal/config"
    "github.com/ajora/auth-service/internal/handlers"
    "github.com/ajora/auth-service/internal/middleware"
    "github.com/ajora/auth-service/internal/repository"
    "github.com/ajora/auth-service/internal/service"
    "github.com/gorilla/mux"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/redis/go-redis/v9"
    "github.com/rs/cors"
)

func main() {
    cfg := config.Load()
    
    pgConn, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
    if err != nil {
        log.Fatalf("Unable to connect to database: %v", err)
    }
    defer pgConn.Close()

    rdb := redis.NewClient(&redis.Options{
        Addr:     cfg.RedisAddr,
        Password: cfg.RedisPassword,
        DB:       cfg.RedisDB,
    })
    defer rdb.Close()

    userRepo := repository.NewUserRepository(pgConn)
    authRepo := repository.NewAuthRepository(pgConn)
    sessionRepo := repository.NewSessionRepository(rdb)

    jwtService := service.NewJWTService(cfg.JWTSecret, cfg.JWTExpiry)
    authService := service.NewAuthService(userRepo, authRepo, sessionRepo, jwtService)
    mfaService := service.NewMFAService()
    
    authHandler := handlers.NewAuthHandler(authService, mfaService)
    
    router := mux.NewRouter()
    
    router.HandleFunc("/api/v1/auth/register", authHandler.Register).Methods("POST")
    router.HandleFunc("/api/v1/auth/login", authHandler.Login).Methods("POST")
    router.HandleFunc("/api/v1/auth/verify-email", authHandler.VerifyEmail).Methods("POST")
    router.HandleFunc("/api/v1/auth/reset-password", authHandler.ResetPassword).Methods("POST")
    router.HandleFunc("/api/v1/auth/refresh", authHandler.RefreshToken).Methods("POST")
    
    auth := router.PathPrefix("/api/v1/auth").Subrouter()
    auth.Use(middleware.Authenticate(jwtService))
    auth.HandleFunc("/logout", authHandler.Logout).Methods("POST")
    auth.HandleFunc("/mfa/enable", authHandler.EnableMFA).Methods("POST")
    auth.HandleFunc("/mfa/disable", authHandler.DisableMFA).Methods("POST")
    auth.HandleFunc("/mfa/verify", authHandler.VerifyMFA).Methods("POST")
    auth.HandleFunc("/change-password", authHandler.ChangePassword).Methods("PUT")
    auth.HandleFunc("/sessions", authHandler.GetSessions).Methods("GET")
    auth.HandleFunc("/sessions/{id}", authHandler.RevokeSession).Methods("DELETE")

    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"https://api.ajora.com"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Request-ID"},
        ExposedHeaders:   []string{"X-Request-ID"},
        AllowCredentials: true,
        MaxAge:           300,
    })

    handler := c.Handler(router)
    handler = middleware.RateLimiter(rdb)(handler)
    handler = middleware.RequestID(handler)
    handler = middleware.Logger(handler)
    handler = middleware.Recovery(handler)

    srv := &http.Server{
        Handler:      handler,
        Addr:         fmt.Sprintf(":%s", cfg.Port),
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    log.Printf("Auth Service starting on port %s", cfg.Port)
    if err := srv.ListenAndServe(); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}

