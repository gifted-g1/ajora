
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

    "github.com/gorilla/mux"
    "github.com/lib/pq"
    "go.uber.org/zap"
)

type User struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    Phone     string    `json:"phone"`
    FirstName string    `json:"first_name"`
    LastName  string    `json:"last_name"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type UserService struct {
    db *sql.DB
    logger *zap.SugaredLogger
}

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()
    sugar := logger.Sugar()

    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        dbURL = "postgres://ajora_admin:password@localhost:5432/ajora?sslmode=disable"
    }

    db, err := sql.Open("postgres", dbURL)
    if err != nil {
        sugar.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()

    service := &UserService{
        db:     db,
        logger: sugar,
    }

    router := mux.NewRouter()

    router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "status": "healthy",
            "service": "user-service",
            "timestamp": time.Now().UTC().Format(time.RFC3339),
        })
    }).Methods("GET")

    router.HandleFunc("/api/v1/users/{id}", service.GetUser).Methods("GET")
    router.HandleFunc("/api/v1/users/{id}", service.UpdateUser).Methods("PUT")
    router.HandleFunc("/api/v1/users/{id}", service.DeleteUser).Methods("DELETE")
    router.HandleFunc("/api/v1/users/{id}/kyc", service.SubmitKYC).Methods("POST")
    router.HandleFunc("/api/v1/users/{id}/kyc", service.GetKYC).Methods("GET")
    router.HandleFunc("/api/v1/users/{id}/wallet", service.CreateWallet).Methods("POST")
    router.HandleFunc("/api/v1/users/{id}/wallet", service.GetWallet).Methods("GET")

    srv := &http.Server{
        Handler: router,
        Addr:    ":8082",
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
    }

    go func() {
        sugar.Info("User Service starting on port 8082")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            sugar.Fatalf("Failed to start server: %v", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    sugar.Info("Shutting down user service...")
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        sugar.Fatalf("Server forced to shutdown: %v", err)
    }

    sugar.Info("User service exited")
}

func (s *UserService) GetUser(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    var user User
    err := s.db.QueryRowContext(r.Context(), `
        SELECT id, email, phone, first_name, last_name, status, created_at, updated_at
        FROM users WHERE id = $1 AND deleted_at IS NULL
    `, id).Scan(&user.ID, &user.Email, &user.Phone, &user.FirstName, &user.LastName,
        &user.Status, &user.CreatedAt, &user.UpdatedAt)

    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "User not found", http.StatusNotFound)
            return
        }
        s.logger.Errorw("Failed to get user", "error", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "data":   user,
    })
}

func (s *UserService) UpdateUser(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    var updates map[string]interface{}
    if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Build update query dynamically
    query := "UPDATE users SET updated_at = NOW()"
    args := []interface{}{}
    i := 1

    for key, value := range updates {
        if key == "id" || key == "created_at" || key == "deleted_at" {
            continue
        }
        query += fmt.Sprintf(", %s = $%d", key, i)
        args = append(args, value)
        i++
    }

    query += fmt.Sprintf(" WHERE id = $%d AND deleted_at IS NULL RETURNING id", i)
    args = append(args, id)

    var returnedID string
    err := s.db.QueryRowContext(r.Context(), query, args...).Scan(&returnedID)
    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "User not found", http.StatusNotFound)
            return
        }
        s.logger.Errorw("Failed to update user", "error", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "message": "User updated successfully",
    })
}

func (s *UserService) DeleteUser(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    result, err := s.db.ExecContext(r.Context(), `
        UPDATE users SET deleted_at = NOW(), status = DELETED 
        WHERE id = $1 AND deleted_at IS NULL
    `, id)

    if err != nil {
        s.logger.Errorw("Failed to delete user", "error", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    rows, _ := result.RowsAffected()
    if rows == 0 {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "message": "User deleted successfully",
    })
}

func (s *UserService) SubmitKYC(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    var kycData map[string]interface{}
    if err := json.NewDecoder(r.Body).Decode(&kycData); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    kycJSON, err := json.Marshal(kycData)
    if err != nil {
        http.Error(w, "Invalid KYC data", http.StatusBadRequest)
        return
    }

    _, err = s.db.ExecContext(r.Context(), `
        UPDATE users SET kyc_data = $1, kyc_status = PENDING 
        WHERE id = $2 AND deleted_at IS NULL
    `, string(kycJSON), id)

    if err != nil {
        s.logger.Errorw("Failed to submit KYC", "error", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "message": "KYC documents submitted successfully",
        "kyc_status": "PENDING",
    })
}

func (s *UserService) GetKYC(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    var kycStatus string
    var kycData string
    err := s.db.QueryRowContext(r.Context(), `
        SELECT kyc_status, COALESCE(kyc_data, {}) FROM users WHERE id = $1 AND deleted_at IS NULL
    `, id).Scan(&kycStatus, &kycData)

    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "User not found", http.StatusNotFound)
            return
        }
        s.logger.Errorw("Failed to get KYC", "error", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "data": map[string]interface{}{
            "kyc_status": kycStatus,
            "kyc_data":   json.RawMessage(kycData),
        },
    })
}

func (s *UserService) CreateWallet(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    userID := vars["id"]

    // Generate wallet address (simplified)
    address := "0x" + fmt.Sprintf("%x", time.Now().UnixNano())

    var walletID string
    err := s.db.QueryRowContext(r.Context(), `
        INSERT INTO wallets (user_id, address, public_key, encrypted_private_key, created_at)
        VALUES ($1, $2, $3, $4, NOW())
        RETURNING id
    `, userID, address, address, "encrypted_key").Scan(&walletID)

    if err != nil {
        s.logger.Errorw("Failed to create wallet", "error", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "data": map[string]interface{}{
            "wallet_id": walletID,
            "address":   address,
        },
    })
}

func (s *UserService) GetWallet(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    userID := vars["id"]

    var wallet struct {
        ID      string `json:"id"`
        Address string `json:"address"`
        Balance string `json:"balance"`
    }

    err := s.db.QueryRowContext(r.Context(), `
        SELECT id, address, balance FROM wallets 
        WHERE user_id = $1 AND is_active = true
        ORDER BY created_at DESC LIMIT 1
    `, userID).Scan(&wallet.ID, &wallet.Address, &wallet.Balance)

    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "Wallet not found", http.StatusNotFound)
            return
        }
        s.logger.Errorw("Failed to get wallet", "error", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "data":   wallet,
    })
}

