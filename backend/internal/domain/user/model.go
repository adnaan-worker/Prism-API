package user

import (
	"time"
)

// User 鐢ㄦ埛妯″瀷
type User struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	Username     string         `gorm:"uniqueIndex;not null;size:255" json:"username"`
	Email        string         `gorm:"uniqueIndex;not null;size:255" json:"email"`
	PasswordHash string         `gorm:"not null;size:255" json:"-"`
	Quota        int64          `gorm:"not null;default:10000" json:"quota"`
	UsedQuota    int64          `gorm:"not null;default:0" json:"used_quota"`
	IsAdmin      bool           `gorm:"not null;default:false" json:"is_admin"`
	Status       string         `gorm:"not null;default:'active';size:50" json:"status"`
	LastSignIn   *time.Time     `json:"last_sign_in,omitempty"`
}

// TableName 鎸囧畾琛ㄥ悕
func (User) TableName() string {
	return "users"
}

// IsActive 妫€鏌ョ敤鎴锋槸鍚︽縺娲?
func (u *User) IsActive() bool {
	return u.Status == "active"
}

// HasQuota 妫€鏌ユ槸鍚︽湁瓒冲閰嶉
func (u *User) HasQuota(required int64) bool {
	return u.Quota-u.UsedQuota >= required
}

// RemainingQuota 鍓╀綑閰嶉
func (u *User) RemainingQuota() int64 {
	remaining := u.Quota - u.UsedQuota
	if remaining < 0 {
		return 0
	}
	return remaining
}

// UseQuota 浣跨敤閰嶉
func (u *User) UseQuota(amount int64) bool {
	if !u.HasQuota(amount) {
		return false
	}
	u.UsedQuota += amount
	return true
}

// AddQuota 澧炲姞閰嶉
func (u *User) AddQuota(amount int64) {
	u.Quota += amount
}

// ResetUsedQuota 閲嶇疆宸蹭娇鐢ㄩ厤棰?
func (u *User) ResetUsedQuota() {
	u.UsedQuota = 0
}
