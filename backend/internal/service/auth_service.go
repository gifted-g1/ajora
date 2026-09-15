package service

import (
    "context"
    "errors"
    "time"

    "github.com/ajora/backend/internal/config"
    "github.com/ajora/backend/internal/domain"
    "github.com/ajora/backend/internal/utils"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type AuthService struct {
    db  *gorm.DB
    cfg *config.Config
}

func NewAuthService(db *gorm.DB, cfg *config.Config) *AuthService {
    return &AuthService{db: db, cfg: cfg}
}

type RegisterInput struct {
    FullName string `json:"full_name" binding:"required"`
    Phone    string `json:"phone" binding:"required"`
    Email    string `json:"email"`
    Password string `json:"password" binding:"required,min=8"`
}

type AuthTokens struct {
    AccessToken  string       `json:"access_token"`
    RefreshToken string       `json:"refresh_token"`
    ExpiresIn    int64        `json:"expires_in"`
    User         *domain.User `json:"user"`
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*domain.User, string, error) {
    var existing domain.User
    if err := s.db.WithContext(ctx).Where("phone = ?", in.Phone).First(&existing).Error; err == nil {
        return nil, "", errors.New("phone number already registered")
    }
    hash, err := utils.HashPassword(in.Password)
    if err != nil {
        return nil, "", err
    }
    user := &domain.User{
        ID: uuid.New(), FullName: in.FullName, Phone: in.Phone,
        Email: in.Email, PasswordHash: hash,
    }
    err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(user).Error; err != nil {
            return err
        }
        return tx.Create(&domain.Wallet{ID: uuid.New(), UserID: user.ID}).Error
    })
    if err != nil {
        return nil, "", err
    }
    otpCode := utils.GenerateOTP()
    otp := &domain.OTPCode{
        ID: uuid.New(), Phone: in.Phone, Code: otpCode,
        ExpiresAt: time.Now().Add(10 * time.Minute),
    }
    if err := s.db.WithContext(ctx).Create(otp).Error; err != nil {
        return nil, "", err
    }
    return user, otpCode, nil
}

func (s *AuthService) VerifyOTP(ctx context.Context, phone, code string) (*AuthTokens, error) {
    var otp domain.OTPCode
    err := s.db.WithContext(ctx).
        Where("phone = ? AND code = ? AND used = false AND expires_at > ?", phone, code, time.Now()).
        First(&otp).Error
    if err != nil {
        return nil, errors.New("invalid or expired OTP")
    }
    var user domain.User
    if err := s.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error; err != nil {
        return nil, errors.New("user not found")
    }
    s.db.WithContext(ctx).Model(&otp).Update("used", true)
    s.db.WithContext(ctx).Model(&user).Update("phone_verified", true)
    access, _ := utils.GenerateAccessToken(user.ID, user.Phone, s.cfg.JWTSecret, s.cfg.AccessTTL)
    refresh, _ := utils.GenerateAccessToken(user.ID, user.Phone, s.cfg.JWTSecret, s.cfg.RefreshTTL)
    return &AuthTokens{
        AccessToken: access, RefreshToken: refresh,
        ExpiresIn: int64(s.cfg.AccessTTL.Seconds()), User: &user,
    }, nil
}

func (s *AuthService) Login(ctx context.Context, phone, password string) (*AuthTokens, error) {
    var user domain.User
    if err := s.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error; err != nil {
        return nil, errors.New("invalid credentials")
    }
    if !utils.CheckPassword(user.PasswordHash, password) {
        return nil, errors.New("invalid credentials")
    }
    access, _ := utils.GenerateAccessToken(user.ID, user.Phone, s.cfg.JWTSecret, s.cfg.AccessTTL)
    refresh, _ := utils.GenerateAccessToken(user.ID, user.Phone, s.cfg.JWTSecret, s.cfg.RefreshTTL)
    return &AuthTokens{
        AccessToken: access, RefreshToken: refresh,
        ExpiresIn: int64(s.cfg.AccessTTL.Seconds()), User: &user,
    }, nil
}
