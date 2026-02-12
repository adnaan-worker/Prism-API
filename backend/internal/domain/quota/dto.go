package quota

import "time"

// QuotaInfoResponse 配额信息响应
type QuotaInfoResponse struct {
	TotalQuota     int64      `json:"total_quota"`
	UsedQuota      int64      `json:"used_quota"`
	RemainingQuota int64      `json:"remaining_quota"`
	LastSignIn     *time.Time `json:"last_sign_in,omitempty"`
}

// SignInResponse 签到响应
type SignInResponse struct {
	QuotaAwarded   int       `json:"quota_awarded"`
	TotalQuota     int64     `json:"total_quota"`
	RemainingQuota int64     `json:"remaining_quota"`
	SignInDate     time.Time `json:"sign_in_date"`
}

// UsageHistoryRequest 使用历史请求
type UsageHistoryRequest struct {
	Days int `form:"days" binding:"omitempty,min=1,max=90"`
}

// UsageHistoryResponse 使用历史响应
type UsageHistoryResponse struct {
	History []UsageHistoryItem `json:"history"`
	Days    int                `json:"days"`
}

// UsageHistoryItem 使用历史项
type UsageHistoryItem struct {
	Date   string `json:"date"`
	Tokens int64  `json:"tokens"`
}

// DeductQuotaRequest 扣除配额请求
type DeductQuotaRequest struct {
	Amount int64 `json:"amount" binding:"required,min=1"`
}

// CheckQuotaRequest 检查配额请求
type CheckQuotaRequest struct {
	Amount int64 `form:"amount" binding:"required,min=1"`
}

// CheckQuotaResponse 检查配额响应
type CheckQuotaResponse struct {
	HasSufficientQuota bool  `json:"has_sufficient_quota"`
	RemainingQuota     int64 `json:"remaining_quota"`
	RequiredAmount     int64 `json:"required_amount"`
}
