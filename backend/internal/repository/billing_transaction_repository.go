package repository

import (
	"api-aggregator/backend/internal/models"
	"context"
	"time"

	"gorm.io/gorm"
)

type BillingTransactionRepository struct {
	db *gorm.DB
}

func NewBillingTransactionRepository(db *gorm.DB) *BillingTransactionRepository {
	return &BillingTransactionRepository{db: db}
}

// Create creates a new billing transaction
func (r *BillingTransactionRepository) Create(ctx context.Context, transaction *models.BillingTransaction) error {
	return r.db.WithContext(ctx).Create(transaction).Error
}

// FindByID retrieves a transaction by ID
func (r *BillingTransactionRepository) FindByID(ctx context.Context, id uint) (*models.BillingTransaction, error) {
	var transaction models.BillingTransaction
	err := r.db.WithContext(ctx).First(&transaction, id).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// FindByUserID retrieves all transactions for a user
func (r *BillingTransactionRepository) FindByUserID(ctx context.Context, userID uint, limit, offset int) ([]*models.BillingTransaction, error) {
	var transactions []*models.BillingTransaction
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error
	return transactions, err
}

// FindByDateRange retrieves transactions within a date range
func (r *BillingTransactionRepository) FindByDateRange(ctx context.Context, userID uint, startDate, endDate time.Time) ([]*models.BillingTransaction, error) {
	var transactions []*models.BillingTransaction
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, startDate, endDate).
		Order("created_at DESC").
		Find(&transactions).Error
	return transactions, err
}

// GetTotalSpent calculates total spent by user
func (r *BillingTransactionRepository) GetTotalSpent(ctx context.Context, userID uint) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&models.BillingTransaction{}).
		Where("user_id = ? AND type = ? AND status = ?", userID, "charge", "completed").
		Select("COALESCE(SUM(micro_credits), 0)").
		Scan(&total).Error
	return total, err
}

// Update updates a transaction
func (r *BillingTransactionRepository) Update(ctx context.Context, transaction *models.BillingTransaction) error {
	return r.db.WithContext(ctx).Save(transaction).Error
}
