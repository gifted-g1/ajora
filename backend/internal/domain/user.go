package domain

import (
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"
)

type User struct {
    ID            uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
    FullName      string         `gorm:"not null" json:"full_name"`
    Phone         string         `gorm:"uniqueIndex;not null" json:"phone"`
    Email         string         `gorm:"uniqueIndex" json:"email"`
    PasswordHash  string         `gorm:"not null" json:"-"`
    PhoneVerified bool           `gorm:"default:false" json:"phone_verified"`
    KYCStatus     string         `gorm:"default:'UNVERIFIED'" json:"kyc_status"`
    TrustScore    int            `gorm:"default:100" json:"trust_score"`
    IsActive      bool           `gorm:"default:true" json:"is_active"`
    CreatedAt     time.Time      `json:"created_at"`
    UpdatedAt     time.Time      `json:"updated_at"`
    DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string { return "users" }

type OTPCode struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
    Phone     string    `gorm:"index;not null"`
    Code      string    `gorm:"not null"`
    ExpiresAt time.Time `gorm:"not null"`
    Used      bool      `gorm:"default:false"`
    CreatedAt time.Time
}

func (OTPCode) TableName() string { return "otp_codes" }
