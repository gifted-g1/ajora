
package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/ajora/auth-service/internal/auth"
)

type AuthHandler struct {
    service *auth.Service
}

func NewAuthHandler(service *auth.Service) *AuthHandler {
    return &AuthHandler{service: service}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    var req auth.RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    user, err := h.service.Register(r.Context(), &req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "data":   user,
    })
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req auth.LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    resp, err := h.service.Login(r.Context(), &req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusUnauthorized)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "data":   resp,
    })
}

func (h *AuthHandler) VerifyMFA(w http.ResponseWriter, r *http.Request) {
    var req struct {
        UserID string `json:"user_id"`
        Code   string `json:"code"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    resp, err := h.service.VerifyMFA(r.Context(), req.UserID, req.Code)
    if err != nil {
        http.Error(w, err.Error(), http.StatusUnauthorized)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "data":   resp,
    })
}

func (h *AuthHandler) EnableMFA(w http.ResponseWriter, r *http.Request) {
    userID := r.Header.Get("X-User-ID")
    if userID == "" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    resp, err := h.service.EnableMFA(r.Context(), userID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "data":   resp,
    })
}

func (h *AuthHandler) DisableMFA(w http.ResponseWriter, r *http.Request) {
    userID := r.Header.Get("X-User-ID")
    if userID == "" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    if err := h.service.DisableMFA(r.Context(), userID); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "message": "MFA disabled successfully",
    })
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
    // Implementation
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "message": "Token refreshed",
    })
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
    // Implementation
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "message": "Logged out successfully",
    })
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
    // Implementation
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "message": "Email verified",
    })
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
    // Implementation
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "message": "Password reset email sent",
    })
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
    // Implementation
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "message": "Password changed successfully",
    })
}

func (h *AuthHandler) GetSessions(w http.ResponseWriter, r *http.Request) {
    // Implementation
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "data":   []interface{}{},
    })
}

func (h *AuthHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
    // Implementation
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "message": "Session revoked",
    })
}

