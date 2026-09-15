package postgres

import (
    "context"

    "github.com/ajora/backend/internal/domain"
    "github.com/google/uuid"
    "gorm.io/gorm"
    "gorm.io/gorm/clause"
)

type AjoRepo struct{ db *gorm.DB }

func NewAjoRepo(db *gorm.DB) *AjoRepo { return &AjoRepo{db: db} }

func (r *AjoRepo) FindByInviteCode(ctx context.Context, code string) (*domain.Ajo, error) {
    var a domain.Ajo
    err := r.db.WithContext(ctx).Where("invite_code = ?", code).First(&a).Error
    return &a, err
}

func (r *AjoRepo) GetRound(ctx context.Context, tx *gorm.DB, ajoID uuid.UUID, roundNum int) (*domain.AjoRound, error) {
    var rd domain.AjoRound
    err := tx.WithContext(ctx).
        Clauses(clause.Locking{Strength: "UPDATE"}).
        Where("ajo_id = ? AND round_number = ?", ajoID, roundNum).
        First(&rd).Error
    return &rd, err
}

func (r *AjoRepo) CreateRound(ctx context.Context, tx *gorm.DB, round *domain.AjoRound) error {
    return tx.WithContext(ctx).Create(round).Error
}

func (r *AjoRepo) UpdateRound(ctx context.Context, tx *gorm.DB, round *domain.AjoRound) error {
    return tx.WithContext(ctx).Save(round).Error
}

func (r *AjoRepo) Update(ctx context.Context, tx *gorm.DB, a *domain.Ajo) error {
    return tx.WithContext(ctx).Save(a).Error
}
