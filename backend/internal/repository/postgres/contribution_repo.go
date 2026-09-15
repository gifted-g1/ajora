package postgres

import (
    "context"
    "time"

    "github.com/ajora/backend/internal/domain"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type ContributionRepo struct{ db *gorm.DB }

func NewContributionRepo(db *gorm.DB) *ContributionRepo { return &ContributionRepo{db: db} }

func (r *ContributionRepo) Create(ctx context.Context, tx *gorm.DB, c *domain.Contribution) error {
    return tx.WithContext(ctx).Create(c).Error
}

func (r *ContributionRepo) CountConfirmedForRound(ctx context.Context, tx *gorm.DB, roundID uuid.UUID) (int64, error) {
    var count int64
    err := tx.WithContext(ctx).Model(&domain.Contribution{}).
        Where("round_id = ? AND status = ?", roundID, domain.ContributionConfirmed).
        Count(&count).Error
    return count, err
}

func (r *ContributionRepo) UpdateStatus(ctx context.Context, tx *gorm.DB, id uuid.UUID, status domain.ContributionStatus) error {
    now := time.Now()
    return tx.WithContext(ctx).Model(&domain.Contribution{}).
        Where("id = ?", id).
        Updates(map[string]interface{}{
            "status": status, "paid_at": &now, "updated_at": now,
        }).Error
}
