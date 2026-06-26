
package auth

import (
    "context"
    "crypto/rand"
    "encoding/base32"
    "errors"
    "time"

    "github.com/ajora/auth-service/internal/db"
    "github.com/ajora/auth-service/internal/models"
    "github.com/golang-jwt/jwt/v5"
    "github.com/pquerna/otp/totp"
    "golang.org/x/crypto/bcrypt"
)

type Service struct {
    db          *sql.DB
    redis       *redis.Client
    config      *config.Config
}

func NewService(db *sql.DB, redis *redis.Client, config *config.Config) *Service {
    return &Service{
        db:     db,
        redis:  redis,
        config: config,
    }
}

func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*models.User, error) {
    // Check if user exists
    var exists bool
    err := s.db.QueryRowContext(ctx,
        "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)",
        req.Email,
    ).Scan(&exists)
    if err != nil {
        return nil, err
    }
    if exists {
        return nil, errors.New("user already exists")
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }

    // Create user
    user := &models.User{
        ID:            generateUUID(),
        Email:         req.Email,
        Phone:         req.Phone,
        PasswordHash:  string(hashedPassword),
        FirstName:     req.FirstName,
        LastName:      req.LastName,
        Role:          "USER",
        Status:        "ACTIVE",
        KYCStatus:     "PENDING",
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
    }

    _, err = s.db.ExecContext(ctx, `
        INSERT INTO users (id, email, phone, password_hash, first_name, last_name, role, status, kyc_status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    `, user.ID, user.Email, user.Phone, user.PasswordHash, user.FirstName, user.LastName,
        user.Role, user.Status, user.KYCStatus, user.CreatedAt, user.UpdatedAt)
    if err != nil {
        return nil, err
    }

    return user, nil
}

func (s *Service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
    var user models.User
    err := s.db.QueryRowContext(ctx, `
        SELECT id, email, phone, password_hash, first_name, last_name, role, status, 
               kyc_status, mfa_enabled, mfa_secret, failed_login_attempts, locked_until
        FROM users WHERE email = $1 AND deleted_at IS NULL
    `, req.Email).Scan(
        &user.ID, &user.Email, &user.Phone, &user.PasswordHash,
        &user.FirstName, &user.LastName, &user.Role, &user.Status,
        &user.KYCStatus, &user.MFAEnabled, &user.MFASecret,
        &user.FailedLoginAttempts, &user.LockedUntil,
    )
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, errors.New("invalid credentials")
        }
        return nil, err
    }

    // Check if account is locked
    if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
        return nil, errors.New("account is locked")
    }

    // Verify password
    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
        // Increment failed attempts
        _, _ = s.db.ExecContext(ctx,
            "UPDATE users SET failed_login_attempts = failed_login_attempts + 1 WHERE id = $1",
            user.ID,
        )

        // Lock account after 5 failures
        var attempts int
        _ = s.db.QueryRowContext(ctx,
            "SELECT failed_login_attempts FROM users WHERE id = $1",
            user.ID,
        ).Scan(&attempts)

        if attempts >= 5 {
            lockedUntil := time.Now().Add(30 * time.Minute)
            _, _ = s.db.ExecContext(ctx,
                "UPDATE users SET locked_until = $1 WHERE id = $2",
                lockedUntil, user.ID,
            )
            return nil, errors.New("account locked due to multiple failed attempts")
        }

        return nil, errors.New("invalid credentials")
    }

    // Reset failed attempts
    _, _ = s.db.ExecContext(ctx,
        "UPDATE users SET failed_login_attempts = 0, last_login = $1 WHERE id = $2",
        time.Now(), user.ID,
    )

    // Check MFA
    if user.MFAEnabled {
        return &LoginResponse{
            UserID:      user.ID,
            RequiresMFA: true,
            TempToken:   s.generateTemporaryToken(user.ID),
        }, nil
    }

    // Generate tokens
    accessToken, err := s.generateAccessToken(&user)
    if err != nil {
        return nil, err
    }

    refreshToken, err := s.generateRefreshToken(&user)
    if err != nil {
        return nil, err
    }

    // Store refresh token in Redis
    err = s.redis.Set(ctx, "refresh:"+refreshToken, user.ID, s.config.RefreshExpiry).Err()
    if err != nil {
        return nil, err
    }

    return &LoginResponse{
        UserID:       user.ID,
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    int(s.config.JWTExpiry.Seconds()),
        RequiresMFA:  false,
    }, nil
}

func (s *Service) VerifyMFA(ctx context.Context, userID, code string) (*LoginResponse, error) {
    var user models.User
    err := s.db.QueryRowContext(ctx, `
        SELECT id, email, mfa_secret FROM users WHERE id = $1
    `, userID).Scan(&user.ID, &user.Email, &user.MFASecret)
    if err != nil {
        return nil, err
    }

    if !totp.Validate(code, user.MFASecret) {
        return nil, errors.New("invalid MFA code")
    }

    accessToken, err := s.generateAccessToken(&user)
    if err != nil {
        return nil, err
    }

    refreshToken, err := s.generateRefreshToken(&user)
    if err != nil {
        return nil, err
    }

    return &LoginResponse{
        UserID:       user.ID,
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    int(s.config.JWTExpiry.Seconds()),
        RequiresMFA:  false,
    }, nil
}

func (s *Service) EnableMFA(ctx context.Context, userID string) (*MFASetupResponse, error) {
    var email string
    err := s.db.QueryRowContext(ctx,
        "SELECT email FROM users WHERE id = $1",
        userID,
    ).Scan(&email)
    if err != nil {
        return nil, err
    }

    key, err := totp.Generate(totp.GenerateOpts{
        Issuer:      s.config.MFAIssuer,
        AccountName: email,
    })
    if err != nil {
        return nil, err
    }

    // Store secret temporarily
    err = s.redis.Set(ctx, "mfa_setup:"+userID, key.Secret(), 10*time.Minute).Err()
    if err != nil {
        return nil, err
    }

    return &MFASetupResponse{
        Secret:    key.Secret(),
        QRCodeURL: key.URL(),
        BackupKeys: s.generateBackupKeys(),
    }, nil
}

func (s *Service) DisableMFA(ctx context.Context, userID string) error {
    _, err := s.db.ExecContext(ctx,
        "UPDATE users SET mfa_enabled = false, mfa_secret = NULL WHERE id = $1",
        userID,
    )
    return err
}

func (s *Service) generateAccessToken(user *models.User) (string, error) {
    claims := jwt.MapClaims{
        "sub":   user.ID,
        "email": user.Email,
        "role":  user.Role,
        "exp":   time.Now().Add(s.config.JWTExpiry).Unix(),
        "iat":   time.Now().Unix(),
        "iss":   "ajora-auth",
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(s.config.JWTSecret))
}

func (s *Service) generateRefreshToken(user *models.User) (string, error) {
    claims := jwt.MapClaims{
        "sub": user.ID,
        "exp": time.Now().Add(s.config.RefreshExpiry).Unix(),
        "iat": time.Now().Unix(),
        "type": "refresh",
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(s.config.JWTSecret))
}

func (s *Service) generateTemporaryToken(userID string) string {
    // Generate a short-lived token for MFA verification
    claims := jwt.MapClaims{
        "sub": userID,
        "exp": time.Now().Add(5 * time.Minute).Unix(),
        "iat": time.Now().Unix(),
        "type": "temp",
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, _ := token.SignedString([]byte(s.config.JWTSecret))
    return tokenString
}

func (s *Service) generateBackupKeys() []string {
    keys := make([]string, 10)
    for i := 0; i < 10; i++ {
        b := make([]byte, 8)
        rand.Read(b)
        keys[i] = base32.StdEncoding.EncodeToString(b)
    }
    return keys
}

func generateUUID() string {
    b := make([]byte, 16)
    rand.Read(b)
    return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

type RegisterRequest struct {
    Email     string `json:"email"`
    Phone     string `json:"phone"`
    Password  string `json:"password"`
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
}

type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
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

