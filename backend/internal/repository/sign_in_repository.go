package repository

import (
	"api-aggregator/backend/internal/models"
	"context"
	"time"

	"gorm.io/gorm"
)

type SignInRepository struct {
	db *gorm.DB
}

func NewSignInRepository(db *gorm.DB) *SignInRepository {
	return &SignInRepository{db: db}
}

// Create creates a new sign-in record
func (r *SignInRepository) Create(ctx context.Context, record *models.SignInRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// HasSignedInToday checks if a user has signed in today
func (r *SignInRepository) HasSignedInToday(ctx context.Context, userID uint) (bool, error) {
	var count int64
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	err := r.db.WithContext(ctx).
		Model(&models.SignInRecord{}).
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, today, tomorrow).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetTodaySignIn gets today's sign-in record for a user
func (r *SignInRepository) GetTodaySignIn(ctx context.Context, userID uint) (*models.SignInRecord, error) {
	var record models.SignInRecord
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, today, tomorrow).
		First(&record).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &record, nil
}
