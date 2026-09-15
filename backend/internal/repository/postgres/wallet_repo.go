package postgres

import (
    "context"
    "errors"
    "time"

    "github.com/ajora/backend/internal/domain"
    "github.com/google/uuid"
    "gorm.io/gorm"
    "gorm.io/gorm/clause"
)

type WalletRepo struct{ db *gorm.DB }

func NewWalletRepo(db *gorm.DB) *WalletRepo { return &WalletRepo{db: db} }

func (r *WalletRepo) GetOrCreate(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error) {
    var w domain.Wallet
    err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&w).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        w = domain.Wallet{ID: uuid.New(), UserID: userID}
        if err := r.db.WithContext(ctx).Create(&w).Error; err != nil {
            return nil, err
        }
        return &w, nil
    }
    return &w, err
}

func (r *WalletRepo) Debit(ctx context.Context, tx *gorm.DB, walletID uuid.UUID, amount int64,
    txType domain.TransactionType, ref, idemKey string, ajoID, roundID *uuid.UUID) (*domain.WalletTransaction, error) {

    var w domain.Wallet
    if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
        Where("id = ?", walletID).First(&w).Error; err != nil {
        return nil, err
    }
    if w.AvailableKobo < amount {
        return nil, errors.New("insufficient balance")
    }
    before := w.AvailableKobo
    w.AvailableKobo -= amount
    w.TotalOutKobo += amount
    if err := tx.WithContext(ctx).Model(&w).Updates(map[string]interface{}{
        "available_kobo": w.AvailableKobo,
        "total_out_kobo": w.TotalOutKobo,
        "updated_at":     time.Now(),
    }).Error; err != nil {
        return nil, err
    }
    entry := domain.WalletTransaction{
        ID: uuid.New(), WalletID: walletID, UserID: w.UserID,
        Type: txType, AmountKobo: -amount, BalanceBefore: before, BalanceAfter: w.AvailableKobo,
        Reference: ref, IdempotencyKey: idemKey, AjoID: ajoID, RoundID: roundID, CreatedAt: time.Now(),
    }
    if err := tx.WithContext(ctx).Create(&entry).Error; err != nil {
        return nil, err
    }
    return &entry, nil
}

func (r *WalletRepo) Credit(ctx context.Context, tx *gorm.DB, walletID uuid.UUID, amount int64,
    txType domain.TransactionType, ref, idemKey string, ajoID, roundID *uuid.UUID) (*domain.WalletTransaction, error) {

    var w domain.Wallet
    if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
        Where("id = ?", walletID).First(&w).Error; err != nil {
        return nil, err
    }
    before := w.AvailableKobo
    w.AvailableKobo += amount
    w.TotalInKobo += amount
    if err := tx.WithContext(ctx).Model(&w).Updates(map[string]interface{}{
        "available_kobo": w.AvailableKobo,
        "total_in_kobo":  w.TotalInKobo,
        "updated_at":     time.Now(),
    }).Error; err != nil {
        return nil, err
    }
    entry := domain.WalletTransaction{
        ID: uuid.New(), WalletID: walletID, UserID: w.UserID,
        Type: txType, AmountKobo: amount, BalanceBefore: before, BalanceAfter: w.AvailableKobo,
        Reference: ref, IdempotencyKey: idemKey, AjoID: ajoID, RoundID: roundID, CreatedAt: time.Now(),
    }
    if err := tx.WithContext(ctx).Create(&entry).Error; err != nil {
        return nil, err
    }
    return &entry, nil
}
