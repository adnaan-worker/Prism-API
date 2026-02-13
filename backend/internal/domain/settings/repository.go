package settings

import (
	"context"

	"gorm.io/gorm"
)

// Repository 设置仓储接口
type Repository interface {
	Get(ctx context.Context, key string) (*Setting, error)
	Set(ctx context.Context, key, value, settingType string) error
	GetMultiple(ctx context.Context, keys []string) (map[string]*Setting, error)
	SetMultiple(ctx context.Context, settings map[string]string) error
	Delete(ctx context.Context, key string) error
	FindAll(ctx context.Context) ([]*Setting, error)
}

type repository struct {
	db *gorm.DB
}

// NewRepository 创建设置仓储实例
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Get 获取设置
func (r *repository) Get(ctx context.Context, key string) (*Setting, error) {
	var setting Setting
	err := r.db.WithContext(ctx).Where("\"key\" = ?", key).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

// Set 设置值
func (r *repository) Set(ctx context.Context, key, value, settingType string) error {
	setting := &Setting{
		Key:   key,
		Value: value,
		Type:  settingType,
	}

	return r.db.WithContext(ctx).
		Where("\"key\" = ?", key).
		Assign(map[string]interface{}{
			"value": value,
			"type":  settingType,
		}).
		FirstOrCreate(setting).Error
}

// GetMultiple 获取多个设置
func (r *repository) GetMultiple(ctx context.Context, keys []string) (map[string]*Setting, error) {
	var settings []*Setting
	err := r.db.WithContext(ctx).Where("\"key\" IN ?", keys).Find(&settings).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]*Setting)
	for _, setting := range settings {
		result[setting.Key] = setting
	}
	return result, nil
}

// SetMultiple 设置多个值
func (r *repository) SetMultiple(ctx context.Context, settings map[string]string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for key, value := range settings {
			setting := &Setting{
				Key:   key,
				Value: value,
				Type:  TypeString,
			}

			if err := tx.Where("\"key\" = ?", key).
				Assign(map[string]interface{}{
					"value": value,
				}).
				FirstOrCreate(setting).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// Delete 删除设置
func (r *repository) Delete(ctx context.Context, key string) error {
	return r.db.WithContext(ctx).Where("\"key\" = ? AND is_system = ?", key, false).Delete(&Setting{}).Error
}

// FindAll 查询所有设置
func (r *repository) FindAll(ctx context.Context) ([]*Setting, error) {
	var settings []*Setting
	err := r.db.WithContext(ctx).Find(&settings).Error
	return settings, err
}
