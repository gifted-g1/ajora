
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/google/uuid"
    "github.com/gorilla/mux"
    "github.com/lib/pq"
    "go.uber.org/zap"
)

type Pool struct {
    ID                   string    `json:"id"`
    Name                 string    `json:"name"`
    Description          string    `json:"description"`
    PoolType             string    `json:"pool_type"`
    TotalSlots           int       `json:"total_slots"`
    FilledSlots          int       `json:"filled_slots"`
    ContributionAmount   float64   `json:"contribution_amount"`
    ContributionFrequency string   `json:"contribution_frequency"`
    TotalRounds          int       `json:"total_rounds"`
    CurrentRound         int       `json:"current_round"`
    StartDate            time.Time `json:"start_date"`
    EndDate              time.Time `json:"end_date"`
    InterestRate         float64   `json:"interest_rate"`
    Status               string    `json:"status"`
    CreatedAt            time.Time `json:"created_at"`
    UpdatedAt            time.Time `json:"updated_at"`
    CreatorID            string    `json:"creator_id"`
}

type PoolService struct {
    db     *sql.DB
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

    service := &PoolService{
        db:     db,
        logger: sugar,
    }

    router := mux.NewRouter()

    router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "status": "healthy",
            "service": "pool-service",
            "timestamp": time.Now().UTC().Format(time.RFC3339),
        })
    }).Methods("GET")

    router.HandleFunc("/api/v1/pools", service.CreatePool).Methods("POST")
    router.HandleFunc("/api/v1/pools", service.ListPools).Methods("GET")
    router.HandleFunc("/api/v1/pools/{id}", service.GetPool).Methods("GET")
    router.HandleFunc("/api/v1/pools/{id}", service.UpdatePool).Methods("PUT")
    router.HandleFunc("/api/v1/pools/{id}", service.DeletePool).Methods("DELETE")
    router.HandleFunc("/api/v1/pools/{id}/join", service.JoinPool).Methods("POST")
    router.HandleFunc("/api/v1/pools/{id}/leave", service.LeavePool).Methods("POST")
    router.HandleFunc("/api/v1/pools/{id}/members", service.GetMembers).Methods("GET")
    router.HandleFunc("/api/v1/pools/{id}/rounds", service.GetRounds).Methods("GET")

    srv := &http.Server{
        Handler: router,
        Addr:    ":8083",
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
    }

    go func() {
        sugar.Info("Pool Service starting on port 8083")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            sugar.Fatalf("Failed to start server: %v", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    sugar.Info("Shutting down pool service...")
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        sugar.Fatalf("Server forced to shutdown: %v", err)
    }

    sugar.Info("Pool service exited")
}

func (s *PoolService) CreatePool(w http.ResponseWriter, r *http.Request) {
    var pool Pool
    if err := json.NewDecoder(r.Body).Decode(&pool); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    pool.ID = uuid.New().String()
    pool.Status = "DRAFT"
    pool.CreatedAt = time.Now()
    pool.UpdatedAt = time.Now()
    pool.FilledSlots = 0
    pool.CurrentRound = 0

    _, err := s.db.ExecContext(r.Context(), `
        INSERT INTO savings_pools (
            id, name, description, pool_type, total_slots, filled_slots,
            contribution_amount, contribution_frequency, total_rounds, current_round,
            start_date, end_date, interest_rate, status, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
    `, pool.ID, pool.Name, pool.Description, pool.PoolType, pool.TotalSlots,
        pool.FilledSlots, pool.ContributionAmount, pool.ContributionFrequency,
        pool.TotalRounds, pool.CurrentRound, pool.StartDate, pool.EndDate,
        pool.InterestRate, pool.Status, pool.CreatedAt, pool.UpdatedAt)

    if err != nil {
        s.logger.Errorw("Failed to create pool", "error", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "data":   pool,
    })
}

func (s *PoolService) ListPools(w http.ResponseWriter, r *http.Request) {
    rows, err := s.db.QueryContext(r.Context(), `
        SELECT id, name, description, pool_type, total_slots, filled_slots,
               contribution_amount, contribution_frequency, total_rounds, current_round,
               start_date, end_date, interest_rate, status, created_at, updated_at
        FROM savings_pools WHERE status IN (ACTIVE, DRAFT)
        ORDER BY created_at DESC LIMIT 100
    `)
    if err != nil {
        s.logger.Errorw("Failed to list pools", "error", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var pools []Pool
    for rows.Next() {
        var pool Pool
        err := rows.Scan(&pool.ID, &pool.Name, &pool.Description, &pool.PoolType,
            &pool.TotalSlots, &pool.FilledSlots, &pool.ContributionAmount,
            &pool.ContributionFrequency, &pool.TotalRounds, &pool.CurrentRound,
            &pool.StartDate, &pool.EndDate, &pool.InterestRate, &pool.Status,
            &pool.CreatedAt, &pool.UpdatedAt)
        if err != nil {
            continue
        }
        pools = append(pools, pool)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "data":   pools,
    })
}

func (s *PoolService) GetPool(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    var pool Pool
    err := s.db.QueryRowContext(r.Context(), `
        SELECT id, name, description, pool_type, total_slots, filled_slots,
               contribution_amount, contribution_frequency, total_rounds, current_round,
               start_date, end_date, interest_rate, status, created_at, updated_at
        FROM savings_pools WHERE id = $1
    `, id).Scan(&pool.ID, &pool.Name, &pool.Description, &pool.PoolType,
        &pool.TotalSlots, &pool.FilledSlots, &pool.ContributionAmount,
        &pool.ContributionFrequency, &pool.TotalRounds, &pool.CurrentRound,
        &pool.StartDate, &pool.EndDate, &pool.InterestRate, &pool.Status,
        &pool.CreatedAt, &pool.UpdatedAt)

    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "Pool not found", http.StatusNotFound)
            return
        }
        s.logger.Errorw("Failed to get pool", "error", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "data":   pool,
    })
}

func (s *PoolService) UpdatePool(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    var updates map[string]interface{}
    if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    query := "UPDATE savings_pools SET updated_at = NOW()"
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

    query += fmt.Sprintf(" WHERE id = $%d RETURNING id", i)
    args = append(args, id)

    var returnedID string
    err := s.db.QueryRowContext(r.Context(), query, args...).Scan(&returnedID)
    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "Pool not found", http.StatusNotFound)
            return
        }
        s.logger.Errorw("Failed to update pool", "error", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "message": "Pool updated successfully",
    })
}

func (s *PoolService) DeletePool(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    result, err := s.db.ExecContext(r.Context(), `
        UPDATE savings_pools SET status = CANCELLED, updated_at = NOW()
        WHERE id = $1 AND status != ACTIVE
    `, id)

    if err != nil {
        s.logger.Errorw("Failed to delete pool", "error", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    rows, _ := result.RowsAffected()
    if rows == 0 {
        http.Error(w, "Pool not found or already active", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "message": "Pool cancelled successfully",
    })
}

func (s *PoolService) JoinPool(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    poolID := vars["id"]

    var req struct {
        UserID string `json:"user_id"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Check if pool exists and has slots
    var currentSlots, totalSlots int
    err := s.db.QueryRowContext(r.Context(), `
        SELECT filled_slots, total_slots FROM savings_pools 
        WHERE id = $1 AND status = ACTIVE
    `, poolID).Scan(&currentSlots, &totalSlots)

    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "Pool not found or not active", http.StatusNotFound)
            return
        }
        s.logger.Errorw("Failed to check pool slots", "error", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    if currentSlots >= totalSlots {
        http.Error(w, "Pool is full", http.StatusBadRequest)
        return
    }

    // Check if user already joined
    var exists bool
    err = s.db.QueryRowContext(r.Context(), `
        SELECT EXISTS(SELECT 1 FROM pool_members WHERE pool_id = $1 AND user_id = $2)
    `, poolID, req.UserID).Scan(&exists)

    if err != nil {
        s.logger.Errorw("Failed to check membership", "error", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    if exists {
        http.Error(w, "User already joined this pool", http.StatusBadRequest)
        return
    }

    // Add member
    _, err = s.db.ExecContext(r.Context(), `
        INSERT INTO pool_members (pool_id, user_id, join_date, status)
        VALUES ($1, $2, NOW(), ACTIVE)
    `, poolID, req.UserID)

    if err != nil {
        s.logger.Errorw("Failed to join pool", "error", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // Update filled slots
    _, err = s.db.ExecContext(r.Context(), `
        UPDATE savings_pools SET filled_slots = filled_slots + 1, updated_at = NOW()
        WHERE id = $1
    `, poolID)

    if err != nil {
        s.logger.Errorw("Failed to update filled slots", "error", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "message": "Joined pool successfully",
    })
}

func (s *PoolService) LeavePool(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    poolID := vars["id"]

    var req struct {
        UserID string `json:"user_id"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    result, err := s.db.ExecContext(r.Context(), `
        UPDATE pool_members SET status = INACTIVE
        WHERE pool_id = $1 AND user_id = $2 AND status = ACTIVE
    `, poolID, req.UserID)

    if err != nil {
        s.logger.Errorw("Failed to leave pool", "error", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    rows, _ := result.RowsAffected()
    if rows == 0 {
        http.Error(w, "Member not found", http.StatusNotFound)
        return
    }

    // Update filled slots
    _, err = s.db.ExecContext(r.Context(), `
        UPDATE savings_pools SET filled_slots = filled_slots - 1, updated_at = NOW()
        WHERE id = $1 AND filled_slots > 0
    `, poolID)

    if err != nil {
        s.logger.Errorw("Failed to update filled slots", "error", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "message": "Left pool successfully",
    })
}

func (s *PoolService) GetMembers(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    poolID := vars["id"]

    rows, err := s.db.QueryContext(r.Context(), `
        SELECT user_id, join_date, status, total_contributed, total_payouts
        FROM pool_members WHERE pool_id = $1
        ORDER BY join_date DESC
    `, poolID)

    if err != nil {
        s.logger.Errorw("Failed to get members", "error", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var members []map[string]interface{}
    for rows.Next() {
        var userID string
        var joinDate time.Time
        var status string
        var totalContributed, totalPayouts float64

        err := rows.Scan(&userID, &joinDate, &status, &totalContributed, &totalPayouts)
        if err != nil {
            continue
        }

        members = append(members, map[string]interface{}{
            "user_id": userID,
            "join_date": joinDate,
            "status": status,
            "total_contributed": totalContributed,
            "total_payouts": totalPayouts,
        })
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "data":   members,
    })
}

func (s *PoolService) GetRounds(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    poolID := vars["id"]

    var totalRounds int
    err := s.db.QueryRowContext(r.Context(), `
        SELECT total_rounds FROM savings_pools WHERE id = $1
    `, poolID).Scan(&totalRounds)

    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "Pool not found", http.StatusNotFound)
            return
        }
        s.logger.Errorw("Failed to get rounds", "error", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    rounds := make([]map[string]interface{}, totalRounds)
    for i := 0; i < totalRounds; i++ {
        rounds[i] = map[string]interface{}{
            "round_number": i + 1,
            "status": "PENDING",
            "contributions": 0,
            "total_amount": 0,
        }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "data":   rounds,
    })
}

