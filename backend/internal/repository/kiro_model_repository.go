package repository

import (
	"context"
	"gorm.io/gorm"
)

type KiroModelMapping struct {
	ID          uint   `gorm:"primaryKey"`
	ModelName   string `gorm:"uniqueIndex;not null"`
	KiroModelID string `gorm:"not null"`
	CreatedAt   string
	UpdatedAt   string
}

func (KiroModelMapping) TableName() string {
	return "kiro_model_mappings"
}

type KiroModelRepository struct {
	db *gorm.DB
}

func NewKiroModelRepository(db *gorm.DB) *KiroModelRepository {
	return &KiroModelRepository{db: db}
}

// GetModelMapping 获取模型映射
func (r *KiroModelRepository) GetModelMapping(ctx context.Context, modelName string) (string, error) {
	var mapping KiroModelMapping
	err := r.db.WithContext(ctx).Where("model_name = ?", modelName).First(&mapping).Error
	if err != nil {
		return "", err
	}
	return mapping.KiroModelID, nil
}

// GetAllMappings 获取所有模型映射
func (r *KiroModelRepository) GetAllMappings(ctx context.Context) (map[string]string, error) {
	var mappings []KiroModelMapping
	err := r.db.WithContext(ctx).Find(&mappings).Error
	if err != nil {
		return nil, err
	}
	
	result := make(map[string]string)
	for _, m := range mappings {
		result[m.ModelName] = m.KiroModelID
	}
	return result, nil
}
