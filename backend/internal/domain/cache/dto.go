package cache

import "time"

// GetCacheStatsRequest 获取缓存统计请求
type GetCacheStatsRequest struct {
	UserID uint `form:"user_id" binding:"omitempty"`
}

// CacheStatsResponse 缓存统计响应
type CacheStatsResponse struct {
	TotalHits    int64 `json:"total_hits"`
	TokensSaved  int64 `json:"tokens_saved"`
	CacheEntries int64 `json:"cache_entries"`
}

// GetCacheListRequest 获取缓存列表请求
type GetCacheListRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	UserID   *uint  `form:"user_id" binding:"omitempty"`
	Model    string `form:"model" binding:"omitempty"`
}

// CacheResponse 缓存响应
type CacheResponse struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	CacheKey    string    `json:"cache_key"`
	QueryText   string    `json:"query_text"`
	Model       string    `json:"model"`
	TokensSaved int       `json:"tokens_saved"`
	HitCount    int       `json:"hit_count"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// CacheListResponse 缓存列表响应
type CacheListResponse struct {
	Caches   []*CacheResponse `json:"caches"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

// CleanExpiredCacheResponse 清理过期缓存响应
type CleanExpiredCacheResponse struct {
	Deleted int64  `json:"deleted"`
	Message string `json:"message"`
}

// ClearUserCacheResponse 清除用户缓存响应
type ClearUserCacheResponse struct {
	Deleted int64  `json:"deleted"`
	Message string `json:"message"`
}

// ToResponse 转换为响应对象
func (c *RequestCache) ToResponse() *CacheResponse {
	return &CacheResponse{
		ID:          c.ID,
		UserID:      c.UserID,
		CacheKey:    c.CacheKey,
		QueryText:   c.QueryText,
		Model:       c.Model,
		TokensSaved: c.TokensSaved,
		HitCount:    c.HitCount,
		ExpiresAt:   c.ExpiresAt,
		CreatedAt:   c.CreatedAt,
	}
}

// ToResponseList 批量转换为响应对象
func ToResponseList(caches []*RequestCache) []*CacheResponse {
	responses := make([]*CacheResponse, len(caches))
	for i, cache := range caches {
		responses[i] = cache.ToResponse()
	}
	return responses
}
