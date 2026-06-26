
package service

import (
    "context"
    "crypto/rand"
    "encoding/base32"
    "errors"
    "fmt"
    "time"

    "github.com/ajora/auth-service/internal/models"
    "github.com/ajora/auth-service/internal/repository"
    "github.com/golang-jwt/jwt/v5"
    "github.com/pquerna/otp/totp"
    "golang.org/x/crypto/bcrypt"
)

type AuthService struct {
    userRepo    repository.UserRepository
    authRepo    repository.AuthRepository
    sessionRepo repository.SessionRepository
    jwtService  *JWTService
}

func NewAuthService(
    userRepo repository.UserRepository,
    authRepo repository.AuthRepository,
    sessionRepo repository.SessionRepository,
    jwtService *JWTService,
) *AuthService {
    return &AuthService{
        userRepo:    userRepo,
        authRepo:    authRepo,
        sessionRepo: sessionRepo,
        jwtService:  jwtService,
    }
}

func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*models.User, error) {
    existing, err := s.userRepo.FindByEmail(ctx, req.Email)
    if err == nil && existing != nil {
        return nil, errors.New("user already exists")
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, fmt.Errorf("failed to hash password: %w", err)
    }

    user := &models.User{
        Email:        req.Email,
        Phone:        req.Phone,
        PasswordHash: string(hashedPassword),
        FirstName:    req.FirstName,
        LastName:     req.LastName,
        Role:         "USER",
        Status:       "PENDING",
        KYCStatus:    "PENDING",
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
    }

    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }

    return user, nil
}

func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
    user, err := s.userRepo.FindByEmail(ctx, req.Email)
    if err != nil {
        return nil, errors.New("invalid credentials")
    }

    if user.Status == "LOCKED" {
        if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
            return nil, errors.New("account locked")
        }
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
        s.userRepo.IncrementFailedLogin(ctx, user.ID)
        if user.FailedLoginAttempts >= 5 {
            lockedUntil := time.Now().Add(30 * time.Minute)
            s.userRepo.LockUser(ctx, user.ID, &lockedUntil)
        }
        return nil, errors.New("invalid credentials")
    }

    s.userRepo.ResetFailedLogin(ctx, user.ID)

    if user.MFAEnabled {
        return &LoginResponse{
            UserID:      user.ID,
            RequiresMFA: true,
            TempToken:   s.jwtService.GenerateTemporaryToken(user.ID),
        }, nil
    }

    accessToken, err := s.jwtService.GenerateAccessToken(user)
    if err != nil {
        return nil, fmt.Errorf("failed to generate access token: %w", err)
    }

    refreshToken, err := s.jwtService.GenerateRefreshToken(user)
    if err != nil {
        return nil, fmt.Errorf("failed to generate refresh token: %w", err)
    }

    session := &models.Session{
        UserID:       user.ID,
        RefreshToken: refreshToken,
        ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
        CreatedAt:    time.Now(),
        IPAddress:    req.IPAddress,
        UserAgent:    req.UserAgent,
    }

    if err := s.sessionRepo.Create(ctx, session); err != nil {
        return nil, fmt.Errorf("failed to create session: %w", err)
    }

    return &LoginResponse{
        UserID:       user.ID,
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    3600,
        RequiresMFA:  false,
    }, nil
}

func (s *AuthService) VerifyMFA(ctx context.Context, userID string, code string) (*LoginResponse, error) {
    user, err := s.userRepo.FindByID(ctx, userID)
    if err != nil {
        return nil, errors.New("user not found")
    }

    if !user.MFAEnabled {
        return nil, errors.New("MFA not enabled")
    }

    valid := totp.Validate(code, user.MFASecret)
    if !valid {
        return nil, errors.New("invalid MFA code")
    }

    accessToken, err := s.jwtService.GenerateAccessToken(user)
    if err != nil {
        return nil, fmt.Errorf("failed to generate access token: %w", err)
    }

    refreshToken, err := s.jwtService.GenerateRefreshToken(user)
    if err != nil {
        return nil, fmt.Errorf("failed to generate refresh token: %w", err)
    }

    return &LoginResponse{
        UserID:       user.ID,
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    3600,
        RequiresMFA:  false,
    }, nil
}

func (s *AuthService) EnableMFA(ctx context.Context, userID string) (*MFASetupResponse, error) {
    user, err := s.userRepo.FindByID(ctx, userID)
    if err != nil {
        return nil, errors.New("user not found")
    }

    secret, err := totp.Generate(totp.GenerateOpts{
        Issuer:      "Ajora",
        AccountName: user.Email,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to generate TOTP secret: %w", err)
    }

    return &MFASetupResponse{
        Secret:     secret.Secret(),
        QRCodeURL:  secret.URL(),
        BackupKeys: s.generateBackupKeys(),
    }, nil
}

func (s *AuthService) generateBackupKeys() []string {
    keys := make([]string, 10)
    for i := 0; i < 10; i++ {
        b := make([]byte, 8)
        rand.Read(b)
        keys[i] = base32.StdEncoding.EncodeToString(b)
    }
    return keys
}

type RegisterRequest struct {
    Email     string `json:"email"`
    Phone     string `json:"phone"`
    Password  string `json:"password"`
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
}

type LoginRequest struct {
    Email     string `json:"email"`
    Password  string `json:"password"`
    IPAddress string `json:"ip_address"`
    UserAgent string `json:"user_agent"`
}

type LoginResponse struct {
    UserID       string `json:"user_id"`
    AccessToken  string `json:"access_token,omitempty"`
    RefreshToken string `json:"refresh_token,omitempty"`
    ExpiresIn    int    `json:"expires_in"`
    RequiresMFA  bool   `json:"requires_mfa"`
    TempToken    string `json:"temp_token,omitempty"`
}

type MFASetupResponse struct {
    Secret     string   `json:"secret"`
    QRCodeURL  string   `json:"qr_code_url"`
    BackupKeys []string `json:"backup_keys"`
}

