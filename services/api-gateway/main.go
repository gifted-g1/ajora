
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gorilla/mux"
    "github.com/gorilla/websocket"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/rs/cors"
    "go.uber.org/zap"
)

var (
    upgrader = websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool { return true },
    }
    clients = make(map[*websocket.Conn]bool)
    broadcast = make(chan []byte)
)

type Config struct {
    Port            string
    AuthServiceURL  string
    UserServiceURL  string
    PoolServiceURL  string
    ContributionURL string
    NotificationURL string
    BlockchainURL   string
    ReputationURL   string
    RateLimit       int
    Timeout         time.Duration
}

func loadConfig() *Config {
    return &Config{
        Port:            getEnv("PORT", "8080"),
        AuthServiceURL:  getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
        UserServiceURL:  getEnv("USER_SERVICE_URL", "http://localhost:8082"),
        PoolServiceURL:  getEnv("POOL_SERVICE_URL", "http://localhost:8083"),
        ContributionURL: getEnv("CONTRIBUTION_SERVICE_URL", "http://localhost:8084"),
        NotificationURL: getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8085"),
        BlockchainURL:   getEnv("BLOCKCHAIN_SERVICE_URL", "http://localhost:8086"),
        ReputationURL:   getEnv("REPUTATION_SERVICE_URL", "http://localhost:8087"),
        RateLimit:       100,
        Timeout:         30 * time.Second,
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()
    sugar := logger.Sugar()

    config := loadConfig()

    router := mux.NewRouter()

    // Health check
    router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{
            "status": "healthy",
            "service": "api-gateway",
            "version": "1.0.0",
            "timestamp": time.Now().UTC().Format(time.RFC3339),
        })
    }).Methods("GET")

    // Metrics
    router.Handle("/metrics", promhttp.Handler())

    // WebSocket
    router.HandleFunc("/ws", handleWebSocket)

    // API Routes
    api := router.PathPrefix("/api/v1").Subrouter()

    // Auth routes
    authRouter := api.PathPrefix("/auth").Subrouter()
    authRouter.HandleFunc("/register", proxyHandler(config.AuthServiceURL+"/api/v1/auth/register")).Methods("POST")
    authRouter.HandleFunc("/login", proxyHandler(config.AuthServiceURL+"/api/v1/auth/login")).Methods("POST")
    authRouter.HandleFunc("/refresh", proxyHandler(config.AuthServiceURL+"/api/v1/auth/refresh")).Methods("POST")
    authRouter.HandleFunc("/logout", authMiddleware(proxyHandler(config.AuthServiceURL+"/api/v1/auth/logout"))).Methods("POST")
    authRouter.HandleFunc("/verify-email", proxyHandler(config.AuthServiceURL+"/api/v1/auth/verify-email")).Methods("POST")
    authRouter.HandleFunc("/reset-password", proxyHandler(config.AuthServiceURL+"/api/v1/auth/reset-password")).Methods("POST")
    authRouter.HandleFunc("/change-password", authMiddleware(proxyHandler(config.AuthServiceURL+"/api/v1/auth/change-password"))).Methods("PUT")
    authRouter.HandleFunc("/mfa/enable", authMiddleware(proxyHandler(config.AuthServiceURL+"/api/v1/auth/mfa/enable"))).Methods("POST")
    authRouter.HandleFunc("/mfa/verify", authMiddleware(proxyHandler(config.AuthServiceURL+"/api/v1/auth/mfa/verify"))).Methods("POST")
    authRouter.HandleFunc("/mfa/disable", authMiddleware(proxyHandler(config.AuthServiceURL+"/api/v1/auth/mfa/disable"))).Methods("POST")
    authRouter.HandleFunc("/sessions", authMiddleware(proxyHandler(config.AuthServiceURL+"/api/v1/auth/sessions"))).Methods("GET")
    authRouter.HandleFunc("/sessions/{id}", authMiddleware(proxyHandler(config.AuthServiceURL+"/api/v1/auth/sessions/{id}"))).Methods("DELETE")

    // User routes
    userRouter := api.PathPrefix("/users").Subrouter()
    userRouter.HandleFunc("/{id}", authMiddleware(proxyHandler(config.UserServiceURL+"/api/v1/users/{id}"))).Methods("GET")
    userRouter.HandleFunc("/{id}", authMiddleware(proxyHandler(config.UserServiceURL+"/api/v1/users/{id}"))).Methods("PUT")
    userRouter.HandleFunc("/{id}", authMiddleware(proxyHandler(config.UserServiceURL+"/api/v1/users/{id}"))).Methods("DELETE")
    userRouter.HandleFunc("/{id}/kyc", authMiddleware(proxyHandler(config.UserServiceURL+"/api/v1/users/{id}/kyc"))).Methods("POST")
    userRouter.HandleFunc("/{id}/kyc", authMiddleware(proxyHandler(config.UserServiceURL+"/api/v1/users/{id}/kyc"))).Methods("GET")
    userRouter.HandleFunc("/{id}/wallet", authMiddleware(proxyHandler(config.UserServiceURL+"/api/v1/users/{id}/wallet"))).Methods("POST")
    userRouter.HandleFunc("/{id}/wallet", authMiddleware(proxyHandler(config.UserServiceURL+"/api/v1/users/{id}/wallet"))).Methods("GET")

    // Pool routes
    poolRouter := api.PathPrefix("/pools").Subrouter()
    poolRouter.HandleFunc("", authMiddleware(proxyHandler(config.PoolServiceURL+"/api/v1/pools"))).Methods("POST")
    poolRouter.HandleFunc("", authMiddleware(proxyHandler(config.PoolServiceURL+"/api/v1/pools"))).Methods("GET")
    poolRouter.HandleFunc("/{id}", authMiddleware(proxyHandler(config.PoolServiceURL+"/api/v1/pools/{id}"))).Methods("GET")
    poolRouter.HandleFunc("/{id}", authMiddleware(proxyHandler(config.PoolServiceURL+"/api/v1/pools/{id}"))).Methods("PUT")
    poolRouter.HandleFunc("/{id}", authMiddleware(proxyHandler(config.PoolServiceURL+"/api/v1/pools/{id}"))).Methods("DELETE")
    poolRouter.HandleFunc("/{id}/join", authMiddleware(proxyHandler(config.PoolServiceURL+"/api/v1/pools/{id}/join"))).Methods("POST")
    poolRouter.HandleFunc("/{id}/leave", authMiddleware(proxyHandler(config.PoolServiceURL+"/api/v1/pools/{id}/leave"))).Methods("POST")
    poolRouter.HandleFunc("/{id}/members", authMiddleware(proxyHandler(config.PoolServiceURL+"/api/v1/pools/{id}/members"))).Methods("GET")
    poolRouter.HandleFunc("/{id}/rounds", authMiddleware(proxyHandler(config.PoolServiceURL+"/api/v1/pools/{id}/rounds"))).Methods("GET")

    // Contribution routes
    contribRouter := api.PathPrefix("/contributions").Subrouter()
    contribRouter.HandleFunc("", authMiddleware(proxyHandler(config.ContributionURL+"/api/v1/contributions"))).Methods("POST")
    contribRouter.HandleFunc("/{id}", authMiddleware(proxyHandler(config.ContributionURL+"/api/v1/contributions/{id}"))).Methods("GET")
    contribRouter.HandleFunc("/users/{id}", authMiddleware(proxyHandler(config.ContributionURL+"/api/v1/users/{id}/contributions"))).Methods("GET")
    contribRouter.HandleFunc("/pools/{id}", authMiddleware(proxyHandler(config.ContributionURL+"/api/v1/pools/{id}/contributions"))).Methods("GET")
    contribRouter.HandleFunc("/retry", authMiddleware(proxyHandler(config.ContributionURL+"/api/v1/contributions/retry"))).Methods("POST")

    // Notification routes
    notifRouter := api.PathPrefix("/notifications").Subrouter()
    notifRouter.HandleFunc("/email", authMiddleware(proxyHandler(config.NotificationURL+"/api/v1/notifications/email"))).Methods("POST")
    notifRouter.HandleFunc("/sms", authMiddleware(proxyHandler(config.NotificationURL+"/api/v1/notifications/sms"))).Methods("POST")
    notifRouter.HandleFunc("/push", authMiddleware(proxyHandler(config.NotificationURL+"/api/v1/notifications/push"))).Methods("POST")
    notifRouter.HandleFunc("", authMiddleware(proxyHandler(config.NotificationURL+"/api/v1/notifications"))).Methods("GET")
    notifRouter.HandleFunc("/{id}", authMiddleware(proxyHandler(config.NotificationURL+"/api/v1/notifications/{id}"))).Methods("GET")

    // Blockchain routes
    blockchainRouter := api.PathPrefix("/blockchain").Subrouter()
    blockchainRouter.HandleFunc("/deploy", authMiddleware(proxyHandler(config.BlockchainURL+"/api/v1/blockchain/deploy"))).Methods("POST")
    blockchainRouter.HandleFunc("/transact", authMiddleware(proxyHandler(config.BlockchainURL+"/api/v1/blockchain/transact"))).Methods("POST")
    blockchainRouter.HandleFunc("/tx/{hash}", authMiddleware(proxyHandler(config.BlockchainURL+"/api/v1/blockchain/tx/{hash}"))).Methods("GET")
    blockchainRouter.HandleFunc("/events", authMiddleware(proxyHandler(config.BlockchainURL+"/api/v1/blockchain/events"))).Methods("GET")
    blockchainRouter.HandleFunc("/gas", authMiddleware(proxyHandler(config.BlockchainURL+"/api/v1/blockchain/gas"))).Methods("GET")

    // Reputation routes
    reputationRouter := api.PathPrefix("/reputation").Subrouter()
    reputationRouter.HandleFunc("/users/{id}", authMiddleware(proxyHandler(config.ReputationURL+"/api/v1/reputation/users/{id}"))).Methods("GET")
    reputationRouter.HandleFunc("/top", authMiddleware(proxyHandler(config.ReputationURL+"/api/v1/reputation/top"))).Methods("GET")
    reputationRouter.HandleFunc("/metrics", authMiddleware(proxyHandler(config.ReputationURL+"/api/v1/reputation/metrics"))).Methods("GET")

    // CORS middleware
    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"*"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Request-ID"},
        ExposedHeaders:   []string{"X-Request-ID"},
        AllowCredentials: true,
        MaxAge:           300,
    })

    handler := c.Handler(router)
    handler = rateLimitMiddleware(handler)
    handler = loggingMiddleware(handler, sugar)
    handler = recoveryMiddleware(handler)

    srv := &http.Server{
        Handler:      handler,
        Addr:         ":" + config.Port,
        WriteTimeout: config.Timeout,
        ReadTimeout:  config.Timeout,
        IdleTimeout:  60 * time.Second,
    }

    // Handle WebSocket broadcasts
    go handleBroadcasts()

    // Graceful shutdown
    go func() {
        sugar.Infof("API Gateway starting on port %s", config.Port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            sugar.Fatalf("Failed to start server: %v", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    sugar.Info("Shutting down server...")
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        sugar.Fatalf("Server forced to shutdown: %v", err)
    }

    sugar.Info("Server exited")
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
        return
    }
    defer conn.Close()

    clients[conn] = true

    for {
        _, msg, err := conn.ReadMessage()
        if err != nil {
            delete(clients, conn)
            break
        }
        broadcast <- msg
    }
}

func handleBroadcasts() {
    for {
        msg := <-broadcast
        for client := range clients {
            err := client.WriteMessage(websocket.TextMessage, msg)
            if err != nil {
                client.Close()
                delete(clients, client)
            }
        }
    }
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        // Validate token - forward to auth service or validate locally
        next(w, r)
    }
}

func proxyHandler(targetURL string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Forward request to target service
        // Implementation depends on your proxy library
        w.Header().Set("X-Proxy", "api-gateway")
        // Simple proxy implementation
        http.DefaultServeMux.ServeHTTP(w, r)
    }
}

func rateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Implement rate limiting
        next.ServeHTTP(w, r)
    })
}

func loggingMiddleware(next http.Handler, logger *zap.SugaredLogger) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        logger.Infow("Request",
            "method", r.Method,
            "path", r.URL.Path,
            "remote_addr", r.RemoteAddr,
            "user_agent", r.UserAgent(),
        )
        next.ServeHTTP(w, r)
    })
}

func recoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
                log.Printf("Panic recovered: %v", err)
            }
        }()
        next.ServeHTTP(w, r)
    })
}

