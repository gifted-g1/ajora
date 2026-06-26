
package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/ajora/auth-service/internal/auth"
    "github.com/ajora/auth-service/internal/config"
    "github.com/ajora/auth-service/internal/db"
    "github.com/ajora/auth-service/internal/handlers"
    "github.com/ajora/auth-service/internal/middleware"
    "github.com/gorilla/mux"
    "github.com/rs/cors"
    "go.uber.org/zap"
)

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()
    sugar := logger.Sugar()

    cfg := config.Load()

    // Initialize database
    dbConn, err := db.Connect(cfg.DatabaseURL)
    if err != nil {
        sugar.Fatalf("Failed to connect to database: %v", err)
    }
    defer dbConn.Close()

    // Initialize Redis
    redisClient := db.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)

    // Initialize services
    authService := auth.NewService(dbConn, redisClient, cfg)
    authHandler := handlers.NewAuthHandler(authService)

    router := mux.NewRouter()

    // Health check
    router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "status": "healthy",
            "service": "auth-service",
            "timestamp": time.Now().UTC().Format(time.RFC3339),
        })
    }).Methods("GET")

    // Public routes
    router.HandleFunc("/api/v1/auth/register", authHandler.Register).Methods("POST")
    router.HandleFunc("/api/v1/auth/login", authHandler.Login).Methods("POST")
    router.HandleFunc("/api/v1/auth/refresh", authHandler.Refresh).Methods("POST")
    router.HandleFunc("/api/v1/auth/verify-email", authHandler.VerifyEmail).Methods("POST")
    router.HandleFunc("/api/v1/auth/reset-password", authHandler.ResetPassword).Methods("POST")

    // Protected routes
    authRouter := router.PathPrefix("/api/v1/auth").Subrouter()
    authRouter.Use(middleware.AuthMiddleware(cfg.JWTSecret))
    authRouter.HandleFunc("/logout", authHandler.Logout).Methods("POST")
    authRouter.HandleFunc("/change-password", authHandler.ChangePassword).Methods("PUT")
    authRouter.HandleFunc("/mfa/enable", authHandler.EnableMFA).Methods("POST")
    authRouter.HandleFunc("/mfa/verify", authHandler.VerifyMFA).Methods("POST")
    authRouter.HandleFunc("/mfa/disable", authHandler.DisableMFA).Methods("POST")
    authRouter.HandleFunc("/sessions", authHandler.GetSessions).Methods("GET")
    authRouter.HandleFunc("/sessions/{id}", authHandler.RevokeSession).Methods("DELETE")

    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"*"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Authorization", "Content-Type"},
        AllowCredentials: true,
    })

    handler := c.Handler(router)
    handler = middleware.LoggingMiddleware(handler, sugar)

    srv := &http.Server{
        Handler:      handler,
        Addr:         ":" + cfg.Port,
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    go func() {
        sugar.Infof("Auth Service starting on port %s", cfg.Port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            sugar.Fatalf("Failed to start server: %v", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    sugar.Info("Shutting down auth service...")
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        sugar.Fatalf("Server forced to shutdown: %v", err)
    }

    sugar.Info("Auth service exited")
}

