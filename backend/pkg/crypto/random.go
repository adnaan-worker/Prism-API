package crypto

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateAPIKey 生成API密钥（带 sk- 前缀）
func GenerateAPIKey() (string, error) {
	// 生成32字节随机数
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	
	// 转换为十六进制字符串并添加前缀
	return "sk-" + hex.EncodeToString(bytes), nil
}

// GenerateRandomString 生成指定长度的随机字符串
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

// GenerateRandomBytes 生成指定长度的随机字节
func GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	return bytes, nil
}
