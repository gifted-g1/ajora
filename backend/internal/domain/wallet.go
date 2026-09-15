package domain

import (
    "time"

    "github.com/google/uuid"
)

type Wallet struct {
    ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
    UserID        uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
    AvailableKobo int64     `gorm:"default:0" json:"available_kobo"`
    LockedKobo    int64     `gorm:"default:0" json:"locked_kobo"`
    TotalInKobo   int64     `gorm:"default:0" json:"total_in_kobo"`
    TotalOutKobo  int64     `gorm:"default:0" json:"total_out_kobo"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}

func (Wallet) TableName() string { return "wallets" }

type WalletTransaction struct {
    ID             uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
    WalletID       uuid.UUID       `gorm:"type:uuid;index;not null" json:"wallet_id"`
    UserID         uuid.UUID       `gorm:"type:uuid;index;not null" json:"user_id"`
    Type           TransactionType `gorm:"not null" json:"type"`
    AmountKobo     int64           `gorm:"not null" json:"amount_kobo"`
    BalanceBefore  int64           `gorm:"not null" json:"balance_before_kobo"`
    BalanceAfter   int64           `gorm:"not null" json:"balance_after_kobo"`
    Reference      string          `gorm:"uniqueIndex;not null" json:"reference"`
    IdempotencyKey string          `gorm:"uniqueIndex" json:"idempotency_key"`
    AjoID          *uuid.UUID      `gorm:"type:uuid;index" json:"ajo_id,omitempty"`
    RoundID        *uuid.UUID      `gorm:"type:uuid;index" json:"round_id,omitempty"`
    CreatedAt      time.Time       `gorm:"index" json:"created_at"`
}

func (WalletTransaction) TableName() string { return "wallet_transactions" }

type Contribution struct {
    ID             uuid.UUID          `gorm:"type:uuid;primaryKey" json:"id"`
    AjoID          uuid.UUID          `gorm:"type:uuid;index;not null" json:"ajo_id"`
    RoundID        uuid.UUID          `gorm:"type:uuid;index;not null" json:"round_id"`
    UserID         uuid.UUID          `gorm:"type:uuid;index;not null" json:"user_id"`
    AmountKobo     int64              `gorm:"not null" json:"amount_kobo"`
    Status         ContributionStatus `gorm:"default:'PENDING'" json:"status"`
    IdempotencyKey string             `gorm:"uniqueIndex" json:"idempotency_key"`
    PaidAt         *time.Time         `json:"paid_at"`
    CreatedAt      time.Time          `json:"created_at"`
    UpdatedAt      time.Time          `json:"updated_at"`
}

func (Contribution) TableName() string { return "contributions" }

type Payout struct {
    ID            uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
    AjoID         uuid.UUID    `gorm:"type:uuid;index;not null" json:"ajo_id"`
    RoundID       uuid.UUID    `gorm:"type:uuid;index;not null" json:"round_id"`
    BeneficiaryID uuid.UUID    `gorm:"type:uuid;index;not null" json:"beneficiary_id"`
    AmountKobo    int64        `gorm:"not null" json:"amount_kobo"`
    Status        PayoutStatus `gorm:"default:'SCHEDULED'" json:"status"`
    ClaimedAt     *time.Time   `json:"claimed_at"`
    CreatedAt     time.Time    `json:"created_at"`
    UpdatedAt     time.Time    `json:"updated_at"`
}

func (Payout) TableName() string { return "payouts" }

type AuditLog struct {
    ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
    UserID     uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
    Action     string    `gorm:"not null" json:"action"`
    Resource   string    `json:"resource"`
    ResourceID string    `json:"resource_id"`
    IPAddress  string    `json:"ip_address"`
    CreatedAt  time.Time `gorm:"index" json:"created_at"`
}

func (AuditLog) TableName() string { return "audit_logs" }
