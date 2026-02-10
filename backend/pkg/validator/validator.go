package validator

import (
	"errors"
	"regexp"
)

var (
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrInvalidUsername = errors.New("invalid username format")
	ErrPasswordTooShort = errors.New("password too short")
	ErrInvalidURL      = errors.New("invalid URL format")
)

// 正则表达式
var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,20}$`)
	urlRegex      = regexp.MustCompile(`^https?://[^\s]+$`)
)

// ValidateEmail 验证邮箱格式
func ValidateEmail(email string) error {
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}
	return nil
}

// ValidateUsername 验证用户名格式
func ValidateUsername(username string) error {
	if !usernameRegex.MatchString(username) {
		return ErrInvalidUsername
	}
	return nil
}

// ValidatePassword 验证密码强度
func ValidatePassword(password string) error {
	if len(password) < 6 {
		return ErrPasswordTooShort
	}
	return nil
}

// ValidateURL 验证 URL 格式
func ValidateURL(url string) error {
	if !urlRegex.MatchString(url) {
		return ErrInvalidURL
	}
	return nil
}

// IsPositive 检查数字是否为正数
func IsPositive(n int) bool {
	return n > 0
}

// IsNonNegative 检查数字是否非负
func IsNonNegative(n int) bool {
	return n >= 0
}

// InRange 检查数字是否在范围内
func InRange(n, min, max int) bool {
	return n >= min && n <= max
}
