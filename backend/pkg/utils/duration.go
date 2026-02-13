package utils

import (
	"fmt"
	"time"
)

// FormatDuration 将秒数转换为时间格式字符串 (e.g., 3600 -> "1h", 90 -> "1h30m")
func FormatDuration(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	if seconds < 3600 {
		minutes := seconds / 60
		remaining := seconds % 60
		if remaining == 0 {
			return fmt.Sprintf("%dm", minutes)
		}
		return fmt.Sprintf("%dm%ds", minutes, remaining)
	}
	hours := seconds / 3600
	remaining := seconds % 3600
	if remaining == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	minutes := remaining / 60
	if minutes == 0 {
		return fmt.Sprintf("%dh%ds", hours, remaining)
	}
	return fmt.Sprintf("%dh%dm", hours, minutes)
}

// ParseDuration 将时间格式字符串转换为秒数 (e.g., "24h" -> 86400, "1h30m" -> 5400)
func ParseDuration(s string) (int, error) {
	var total int
	var num int
	var hasDigit bool

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= '0' && c <= '9' {
			num = num*10 + int(c-'0')
			hasDigit = true
		} else if hasDigit {
			switch c {
			case 'h':
				total += num * 3600
			case 'm':
				total += num * 60
			case 's':
				total += num
			default:
				return 0, fmt.Errorf("invalid duration unit: %c", c)
			}
			num = 0
			hasDigit = false
		} else {
			return 0, fmt.Errorf("invalid duration format")
		}
	}

	if hasDigit {
		return 0, fmt.Errorf("missing duration unit")
	}

	if total == 0 {
		return 0, fmt.Errorf("duration cannot be zero")
	}

	return total, nil
}

// ParseDurationToTime 将时间格式字符串转换为 time.Duration
func ParseDurationToTime(s string) (time.Duration, error) {
	seconds, err := ParseDuration(s)
	if err != nil {
		return 0, err
	}
	return time.Duration(seconds) * time.Second, nil
}

// FormatDurationFromTime 将 time.Duration 转换为时间格式字符串
func FormatDurationFromTime(d time.Duration) string {
	return FormatDuration(int(d.Seconds()))
}
