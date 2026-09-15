package domain

import (
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"
)

type Ajo struct {
    ID                 uuid.UUID           `gorm:"type:uuid;primaryKey" json:"id"`
    Name               string              `gorm:"not null" json:"name"`
    Description        string              `json:"description"`
    OrganizerID        uuid.UUID           `gorm:"type:uuid;index;not null" json:"organizer_id"`
    IsPublic           bool                `gorm:"default:false" json:"is_public"`
    InviteCode         string              `gorm:"uniqueIndex;not null" json:"invite_code"`
    ContributionAmount int64               `gorm:"not null" json:"contribution_amount_kobo"`
    Frequency          Frequency           `gorm:"not null" json:"frequency"`
    TotalRounds        int                 `gorm:"not null" json:"total_rounds"`
    MaxMembers         int                 `gorm:"not null" json:"max_members"`
    MinMembers         int                 `gorm:"not null" json:"min_members"`
    PayoutOrder        PayoutOrderStrategy `gorm:"not null" json:"payout_order"`
    GracePeriodHours   int                 `gorm:"default:24" json:"grace_period_hours"`
    LateFeeKobo        int64               `gorm:"default:0" json:"late_fee_kobo"`
    Status             AjoStatus           `gorm:"default:'DRAFT'" json:"status"`
    CurrentRound       int                 `gorm:"default:0" json:"current_round"`
    CreatedAt          time.Time           `json:"created_at"`
    UpdatedAt          time.Time           `json:"updated_at"`
    DeletedAt          gorm.DeletedAt      `gorm:"index" json:"-"`
}

func (Ajo) TableName() string { return "ajos" }

type AjoMember struct {
    ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
    AjoID          uuid.UUID `gorm:"type:uuid;index;not null" json:"ajo_id"`
    UserID         uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
    PayoutPosition int       `gorm:"not null" json:"payout_position"`
    JoinedAt       time.Time `json:"joined_at"`
}

func (AjoMember) TableName() string { return "ajo_members" }

type AjoRound struct {
    ID             uuid.UUID   `gorm:"type:uuid;primaryKey" json:"id"`
    AjoID          uuid.UUID   `gorm:"type:uuid;index;not null" json:"ajo_id"`
    RoundNumber    int         `gorm:"not null" json:"round_number"`
    BeneficiaryID  uuid.UUID   `gorm:"type:uuid;not null" json:"beneficiary_id"`
    Status         RoundStatus `gorm:"default:'UPCOMING'" json:"status"`
    TotalCollected int64       `gorm:"default:0" json:"total_collected_kobo"`
    ExpectedTotal  int64       `gorm:"not null" json:"expected_total_kobo"`
    DueDate        time.Time   `gorm:"not null" json:"due_date"`
    CompletedAt    *time.Time  `json:"completed_at"`
    CreatedAt      time.Time   `json:"created_at"`
    UpdatedAt      time.Time   `json:"updated_at"`
}

func (AjoRound) TableName() string { return "ajo_rounds" }
