package cache

import (
	"time"
)

// RequestCache 请求缓存模型
type RequestCache struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index;not null" json:"user_id"`
	CacheKey    string    `gorm:"type:varchar(32);uniqueIndex;not null" json:"cache_key"`
	QueryText   string    `gorm:"type:text" json:"query_text"`
	Embedding   string    `gorm:"type:text" json:"embedding"`
	Model       string    `gorm:"type:varchar(100);index;not null" json:"model"`
	Request     string    `gorm:"type:text;not null" json:"request"`
	Response    string    `gorm:"type:text;not null" json:"response"`
	TokensSaved int       `gorm:"not null;default:0" json:"tokens_saved"`
	HitCount    int       `gorm:"not null;default:0" json:"hit_count"`
	ExpiresAt   time.Time `gorm:"index;not null" json:"expires_at"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (RequestCache) TableName() string {
	return "request_caches"
}

// IsExpired 检查是否过期
func (c *RequestCache) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// IncrementHitCount 增加命中次数
func (c *RequestCache) IncrementHitCount() {
	c.HitCount++
}

// HasEmbedding 检查是否有向量嵌入
func (c *RequestCache) HasEmbedding() bool {
	return c.Embedding != "" && c.Embedding != "null"
}
