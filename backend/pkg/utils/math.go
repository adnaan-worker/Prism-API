package utils

import "math"

// Min 返回两个整数中的最小值
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max 返回两个整数中的最大值
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// MinInt64 返回两个 int64 中的最小值
func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// MaxInt64 返回两个 int64 中的最大值
func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// Clamp 将值限制在指定范围内
func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// RoundUp 向上取整（用于费用计算）
func RoundUp(value float64) int64 {
	return int64(math.Ceil(value))
}

// RoundDown 向下取整
func RoundDown(value float64) int64 {
	return int64(math.Floor(value))
}

// Round 四舍五入
func Round(value float64) int64 {
	return int64(math.Round(value))
}

// Percentage 计算百分比
func Percentage(value, total float64) float64 {
	if total == 0 {
		return 0
	}
	return (value / total) * 100
}
