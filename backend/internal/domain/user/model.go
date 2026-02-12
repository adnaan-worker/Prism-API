package user

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Username     string         `gorm:"uniqueIndex;not null;size:255" json:"username"`
	Email        string         `gorm:"uniqueIndex;not null;size:255" json:"email"`
	PasswordHash string         `gorm:"not null;size:255" json:"-"`
	Quota        int64          `gorm:"not null;default:10000" json:"quota"`
	UsedQuota    int64          `gorm:"not null;default:0" json:"used_quota"`
	IsAdmin      bool           `gorm:"not null;default:false" json:"is_admin"`
	Status       string         `gorm:"not null;default:'active';size:50" json:"status"`
	LastSignIn   *time.Time     `json:"last_sign_in,omitempty"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// IsActive 检查用户是否激活
func (u *User) IsActive() bool {
	return u.Status == "active"
}

// HasQuota 检查是否有足够配额
func (u *User) HasQuota(required int64) bool {
	return u.Quota-u.UsedQuota >= required
}

// RemainingQuota 剩余配额
func (u *User) RemainingQuota() int64 {
	remaining := u.Quota - u.UsedQuota
	if remaining < 0 {
		return 0
	}
	return remaining
}

// UseQuota 使用配额
func (u *User) UseQuota(amount int64) bool {
	if !u.HasQuota(amount) {
		return false
	}
	u.UsedQuota += amount
	return true
}

// AddQuota 增加配额
func (u *User) AddQuota(amount int64) {
	u.Quota += amount
}

// ResetUsedQuota 重置已使用配额
func (u *User) ResetUsedQuota() {
	u.UsedQuota = 0
}
