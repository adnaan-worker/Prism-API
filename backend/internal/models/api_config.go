package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// StringArray is a custom type for storing string arrays as JSON in PostgreSQL
type StringArray []string

func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return json.Marshal([]string{})
	}
	return json.Marshal(s)
}

func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, s)
}

// JSONMap is a custom type for storing JSON objects in PostgreSQL
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return json.Marshal(map[string]interface{}{})
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = map[string]interface{}{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

type APIConfig struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name      string         `gorm:"not null;size:255" json:"name"`
	Type      string         `gorm:"not null;size:50" json:"type"`
	BaseURL   string         `gorm:"not null;type:text" json:"base_url"`
	APIKey    string         `gorm:"type:text" json:"api_key,omitempty"`
	Models    StringArray    `gorm:"type:jsonb;not null;default:'[]'" json:"models"`
	Headers   JSONMap        `gorm:"type:jsonb" json:"headers,omitempty"`
	IsActive  bool           `gorm:"not null;default:true" json:"is_active"`
	Priority  int            `gorm:"not null;default:100" json:"priority"`
	Weight    int            `gorm:"not null;default:1" json:"weight"`
	MaxRPS    int            `gorm:"not null;default:0" json:"max_rps"`
	Timeout   int            `gorm:"not null;default:30" json:"timeout"`
}

func (APIConfig) TableName() string {
	return "api_configs"
}
