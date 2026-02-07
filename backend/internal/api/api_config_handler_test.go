package api

import (
	"api-aggregator/backend/internal/models"
	"api-aggregator/backend/internal/repository"
	"api-aggregator/backend/internal/service"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *gorm.DB {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	return db
}

func TestBatchDeleteConfigs(t *testing.T) {
	db := setupTestDB(t)

	// Clean up
	db.Exec("DELETE FROM api_configs")

	// Setup
	configRepo := repository.NewAPIConfigRepository(db)
	configService := service.NewAPIConfigService(configRepo)
	handler := NewAPIConfigHandler(configService)

	// Create test configs
	config1 := &models.APIConfig{
		Name:     "Test Config 1",
		Type:     "openai",
		BaseURL:  "https://api.openai.com",
		Models:   []string{"gpt-4"},
		IsActive: true,
		Priority: 100,
		Weight:   1,
		Timeout:  30,
	}
	config2 := &models.APIConfig{
		Name:     "Test Config 2",
		Type:     "anthropic",
		BaseURL:  "https://api.anthropic.com",
		Models:   []string{"claude-3"},
		IsActive: true,
		Priority: 100,
		Weight:   1,
		Timeout:  30,
	}

	if err := configRepo.Create(nil, config1); err != nil {
		t.Fatalf("Failed to create config1: %v", err)
	}
	if err := configRepo.Create(nil, config2); err != nil {
		t.Fatalf("Failed to create config2: %v", err)
	}

	// Test batch delete
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/batch/delete", handler.BatchDeleteConfigs)

	reqBody := map[string]interface{}{
		"ids": []uint{config1.ID, config2.ID},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/batch/delete", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify configs are deleted
	configs, err := configRepo.FindAll(nil)
	if err != nil {
		t.Fatalf("Failed to find configs: %v", err)
	}
	if len(configs) != 0 {
		t.Errorf("Expected 0 configs after deletion, got %d", len(configs))
	}
}

func TestBatchActivateConfigs(t *testing.T) {
	db := setupTestDB(t)

	// Clean up
	db.Exec("DELETE FROM api_configs")

	// Setup
	configRepo := repository.NewAPIConfigRepository(db)
	configService := service.NewAPIConfigService(configRepo)
	handler := NewAPIConfigHandler(configService)

	// Create test configs (inactive)
	config1 := &models.APIConfig{
		Name:     "Test Config 1",
		Type:     "openai",
		BaseURL:  "https://api.openai.com",
		Models:   []string{"gpt-4"},
		IsActive: false,
		Priority: 100,
		Weight:   1,
		Timeout:  30,
	}
	config2 := &models.APIConfig{
		Name:     "Test Config 2",
		Type:     "anthropic",
		BaseURL:  "https://api.anthropic.com",
		Models:   []string{"claude-3"},
		IsActive: false,
		Priority: 100,
		Weight:   1,
		Timeout:  30,
	}

	if err := configRepo.Create(nil, config1); err != nil {
		t.Fatalf("Failed to create config1: %v", err)
	}
	if err := configRepo.Create(nil, config2); err != nil {
		t.Fatalf("Failed to create config2: %v", err)
	}

	// Test batch activate
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/batch/activate", handler.BatchActivateConfigs)

	reqBody := map[string]interface{}{
		"ids": []uint{config1.ID, config2.ID},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/batch/activate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify configs are activated
	updatedConfig1, _ := configRepo.FindByID(nil, config1.ID)
	updatedConfig2, _ := configRepo.FindByID(nil, config2.ID)

	if !updatedConfig1.IsActive {
		t.Errorf("Expected config1 to be active")
	}
	if !updatedConfig2.IsActive {
		t.Errorf("Expected config2 to be active")
	}
}

func TestBatchDeactivateConfigs(t *testing.T) {
	db := setupTestDB(t)

	// Clean up
	db.Exec("DELETE FROM api_configs")

	// Setup
	configRepo := repository.NewAPIConfigRepository(db)
	configService := service.NewAPIConfigService(configRepo)
	handler := NewAPIConfigHandler(configService)

	// Create test configs (active)
	config1 := &models.APIConfig{
		Name:     "Test Config 1",
		Type:     "openai",
		BaseURL:  "https://api.openai.com",
		Models:   []string{"gpt-4"},
		IsActive: true,
		Priority: 100,
		Weight:   1,
		Timeout:  30,
	}
	config2 := &models.APIConfig{
		Name:     "Test Config 2",
		Type:     "anthropic",
		BaseURL:  "https://api.anthropic.com",
		Models:   []string{"claude-3"},
		IsActive: true,
		Priority: 100,
		Weight:   1,
		Timeout:  30,
	}

	if err := configRepo.Create(nil, config1); err != nil {
		t.Fatalf("Failed to create config1: %v", err)
	}
	if err := configRepo.Create(nil, config2); err != nil {
		t.Fatalf("Failed to create config2: %v", err)
	}

	// Test batch deactivate
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/batch/deactivate", handler.BatchDeactivateConfigs)

	reqBody := map[string]interface{}{
		"ids": []uint{config1.ID, config2.ID},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/batch/deactivate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify configs are deactivated
	updatedConfig1, _ := configRepo.FindByID(nil, config1.ID)
	updatedConfig2, _ := configRepo.FindByID(nil, config2.ID)

	if updatedConfig1.IsActive {
		t.Errorf("Expected config1 to be inactive")
	}
	if updatedConfig2.IsActive {
		t.Errorf("Expected config2 to be inactive")
	}
}
