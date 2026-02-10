package models

import (
	"time"
)

// RequestCache 请求缓存模型
type RequestCache struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index;not null" json:"user_id"`
	CacheKey    string    `gorm:"type:varchar(32);uniqueIndex;not null" json:"cache_key"` // MD5 哈希
	QueryText   string    `gorm:"type:text" json:"query_text"`                             // 用于语义匹配的查询文本
	Embedding   string    `gorm:"type:text" json:"embedding"`                              // 向量嵌入（JSON 数组）
	Model       string    `gorm:"type:varchar(100);index;not null" json:"model"`
	Request     string    `gorm:"type:text;not null" json:"request"`       // JSON 序列化的请求
	Response    string    `gorm:"type:text;not null" json:"response"`      // JSON 序列化的响应
	TokensSaved int       `gorm:"not null;default:0" json:"tokens_saved"`  // 该缓存节省的 tokens
	HitCount    int       `gorm:"not null;default:0" json:"hit_count"`     // 缓存命中次数
	ExpiresAt   time.Time `gorm:"index;not null" json:"expires_at"`        // 过期时间
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (RequestCache) TableName() string {
	return "request_caches"
}
