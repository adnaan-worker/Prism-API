package utils

import (
	"encoding/json"
	"fmt"
	"math"
)

// CosineSimilarity 计算余弦相似度
func CosineSimilarity(a, b []float64) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("vectors must have same length")
	}

	if len(a) == 0 {
		return 0, fmt.Errorf("vectors cannot be empty")
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0, nil
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB)), nil
}

// VectorToJSON 将向量转换为 JSON 字符串
func VectorToJSON(vector []float64) (string, error) {
	if vector == nil {
		return "", nil
	}
	data, err := json.Marshal(vector)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// JSONToVector 将 JSON 字符串转换为向量
func JSONToVector(jsonStr string) ([]float64, error) {
	if jsonStr == "" || jsonStr == "null" {
		return nil, nil
	}
	var vector []float64
	err := json.Unmarshal([]byte(jsonStr), &vector)
	if err != nil {
		return nil, err
	}
	return vector, nil
}
