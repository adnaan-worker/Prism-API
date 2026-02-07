package service

import (
	"api-aggregator/backend/internal/models"
	"api-aggregator/backend/internal/repository"
	"context"
	"os"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// getConfigTestDB creates a test database connection
func getConfigTestDB(t *testing.T) *gorm.DB {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "host=localhost user=postgres password=postgres dbname=api_aggregator_test port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	return db
}

// setupConfigTestDB initializes the test database
func setupConfigTestDB(t *testing.T) *gorm.DB {
	db := getConfigTestDB(t)

	// Clean up existing data
	db.Exec("TRUNCATE TABLE api_configs CASCADE")

	// Run migrations
	err := db.AutoMigrate(&models.APIConfig{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

// cleanupConfigTestDB cleans up the test database
func cleanupConfigTestDB(t *testing.T, db *gorm.DB) {
	db.Exec("TRUNCATE TABLE api_configs CASCADE")
}

// Property 7: API配置完整性
// Feature: api-aggregator, Property 7: For any API configuration created by an admin, querying the configuration should return all fields (name, type, base_url, api_key, models) with correct values.
// Validates: Requirements 3.1
func TestProperty_APIConfigIntegrity(t *testing.T) {
	db := setupConfigTestDB(t)
	defer cleanupConfigTestDB(t, db)

	configRepo := repository.NewAPIConfigRepository(db)
	configService := NewAPIConfigService(configRepo)

	properties := gopter.NewProperties(nil)

	// Custom generators
	nameGen := gen.RegexMatch("[a-zA-Z][a-zA-Z0-9 ]{2,30}")
	typeGen := gen.OneConstOf("openai", "anthropic", "gemini", "custom")
	baseURLGen := gen.RegexMatch("https://[a-z]{3,15}\\.com")
	apiKeyGen := gen.RegexMatch("sk-[a-zA-Z0-9]{20,40}")
	modelsGen := gen.SliceOfN(3, gen.RegexMatch("[a-z]{3,10}-[0-9]"))

	properties.Property("API config integrity", prop.ForAll(
		func(name, configType, baseURL, apiKey string, models []string) bool {
			// Clean up before each test
			db.Exec("TRUNCATE TABLE api_configs CASCADE")

			ctx := context.Background()

			// Create config
			req := &CreateConfigRequest{
				Name:     name,
				Type:     configType,
				BaseURL:  baseURL,
				APIKey:   apiKey,
				Models:   models,
				Priority: 100,
				Weight:   1,
				Timeout:  30,
			}

			config, err := configService.CreateConfig(ctx, req)
			if err != nil {
				return false // Config creation should succeed
			}

			// Query the config
			queriedConfig, err := configService.GetConfig(ctx, config.ID)
			if err != nil {
				return false // Query should succeed
			}

			// Verify all fields match
			if queriedConfig.Name != name {
				return false
			}
			if queriedConfig.Type != configType {
				return false
			}
			if queriedConfig.BaseURL != baseURL {
				return false
			}
			if queriedConfig.APIKey != apiKey {
				return false
			}
			if len(queriedConfig.Models) != len(models) {
				return false
			}
			for i, model := range models {
				if queriedConfig.Models[i] != model {
					return false
				}
			}

			return true
		},
		nameGen,
		typeGen,
		baseURLGen,
		apiKeyGen,
		modelsGen,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property 8: 模型数组存储
// Feature: api-aggregator, Property 8: For any API configuration with multiple models, the models field should be stored and retrieved as a JSON array containing all specified models.
// Validates: Requirements 3.2
func TestProperty_ModelArrayStorage(t *testing.T) {
	db := setupConfigTestDB(t)
	defer cleanupConfigTestDB(t, db)

	configRepo := repository.NewAPIConfigRepository(db)
	configService := NewAPIConfigService(configRepo)

	properties := gopter.NewProperties(nil)

	// Generator for multiple models (2-5 models)
	modelsGen := gen.SliceOfN(3, gen.RegexMatch("[a-z]{3,10}-[0-9]"))

	properties.Property("Model array storage", prop.ForAll(
		func(models []string) bool {
			// Clean up before each test
			db.Exec("TRUNCATE TABLE api_configs CASCADE")

			ctx := context.Background()

			// Create config with multiple models
			req := &CreateConfigRequest{
				Name:     "Test Config",
				Type:     "openai",
				BaseURL:  "https://api.openai.com",
				APIKey:   "sk-test",
				Models:   models,
				Priority: 100,
				Weight:   1,
				Timeout:  30,
			}

			config, err := configService.CreateConfig(ctx, req)
			if err != nil {
				return false
			}

			// Query the config
			queriedConfig, err := configService.GetConfig(ctx, config.ID)
			if err != nil {
				return false
			}

			// Verify models array is stored correctly
			if len(queriedConfig.Models) != len(models) {
				return false
			}

			// Verify all models are present in the same order
			for i, model := range models {
				if queriedConfig.Models[i] != model {
					return false
				}
			}

			return true
		},
		modelsGen,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Unit test for CreateConfig
func TestAPIConfigService_CreateConfig(t *testing.T) {
	db := setupConfigTestDB(t)
	defer cleanupConfigTestDB(t, db)

	configRepo := repository.NewAPIConfigRepository(db)
	configService := NewAPIConfigService(configRepo)

	ctx := context.Background()

	// Test successful creation
	req := &CreateConfigRequest{
		Name:     "OpenAI Official",
		Type:     "openai",
		BaseURL:  "https://api.openai.com",
		APIKey:   "sk-test123",
		Models:   []string{"gpt-4", "gpt-3.5-turbo"},
		Priority: 100,
		Weight:   1,
		Timeout:  30,
	}

	config, err := configService.CreateConfig(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if config.Name != req.Name {
		t.Errorf("Expected name %s, got %s", req.Name, config.Name)
	}
	if config.Type != req.Type {
		t.Errorf("Expected type %s, got %s", req.Type, config.Type)
	}
	if !config.IsActive {
		t.Error("Expected config to be active")
	}
	if len(config.Models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(config.Models))
	}

	// Test invalid type
	req2 := &CreateConfigRequest{
		Name:    "Invalid Config",
		Type:    "invalid",
		BaseURL: "https://api.test.com",
		Models:  []string{"model-1"},
	}
	_, err = configService.CreateConfig(ctx, req2)
	if err != ErrInvalidConfig {
		t.Errorf("Expected ErrInvalidConfig, got %v", err)
	}

	// Test empty models array
	req3 := &CreateConfigRequest{
		Name:    "Empty Models",
		Type:    "openai",
		BaseURL: "https://api.test.com",
		Models:  []string{},
	}
	_, err = configService.CreateConfig(ctx, req3)
	if err == nil {
		t.Error("Expected error for empty models array")
	}
}

// Unit test for GetConfig
func TestAPIConfigService_GetConfig(t *testing.T) {
	db := setupConfigTestDB(t)
	defer cleanupConfigTestDB(t, db)

	configRepo := repository.NewAPIConfigRepository(db)
	configService := NewAPIConfigService(configRepo)

	ctx := context.Background()

	// Create a config
	req := &CreateConfigRequest{
		Name:    "Test Config",
		Type:    "openai",
		BaseURL: "https://api.openai.com",
		Models:  []string{"gpt-4"},
	}
	config, err := configService.CreateConfig(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Get the config
	queriedConfig, err := configService.GetConfig(ctx, config.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if queriedConfig.ID != config.ID {
		t.Errorf("Expected ID %d, got %d", config.ID, queriedConfig.ID)
	}

	// Test non-existent config
	_, err = configService.GetConfig(ctx, 99999)
	if err != ErrConfigNotFound {
		t.Errorf("Expected ErrConfigNotFound, got %v", err)
	}
}

// Unit test for UpdateConfig
func TestAPIConfigService_UpdateConfig(t *testing.T) {
	db := setupConfigTestDB(t)
	defer cleanupConfigTestDB(t, db)

	configRepo := repository.NewAPIConfigRepository(db)
	configService := NewAPIConfigService(configRepo)

	ctx := context.Background()

	// Create a config
	req := &CreateConfigRequest{
		Name:    "Original Name",
		Type:    "openai",
		BaseURL: "https://api.openai.com",
		Models:  []string{"gpt-4"},
	}
	config, err := configService.CreateConfig(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Update the config
	updateReq := &UpdateConfigRequest{
		Name:   "Updated Name",
		Models: []string{"gpt-4", "gpt-3.5-turbo"},
	}
	updatedConfig, err := configService.UpdateConfig(ctx, config.ID, updateReq)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if updatedConfig.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got %s", updatedConfig.Name)
	}
	if len(updatedConfig.Models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(updatedConfig.Models))
	}
}

// Unit test for DeleteConfig
func TestAPIConfigService_DeleteConfig(t *testing.T) {
	db := setupConfigTestDB(t)
	defer cleanupConfigTestDB(t, db)

	configRepo := repository.NewAPIConfigRepository(db)
	configService := NewAPIConfigService(configRepo)

	ctx := context.Background()

	// Create a config
	req := &CreateConfigRequest{
		Name:    "Test Config",
		Type:    "openai",
		BaseURL: "https://api.openai.com",
		Models:  []string{"gpt-4"},
	}
	config, err := configService.CreateConfig(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Delete the config
	err = configService.DeleteConfig(ctx, config.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify config is deleted
	_, err = configService.GetConfig(ctx, config.ID)
	if err != ErrConfigNotFound {
		t.Errorf("Expected ErrConfigNotFound, got %v", err)
	}
}

// Unit test for ActivateConfig and DeactivateConfig
func TestAPIConfigService_ActivateDeactivate(t *testing.T) {
	db := setupConfigTestDB(t)
	defer cleanupConfigTestDB(t, db)

	configRepo := repository.NewAPIConfigRepository(db)
	configService := NewAPIConfigService(configRepo)

	ctx := context.Background()

	// Create a config
	req := &CreateConfigRequest{
		Name:    "Test Config",
		Type:    "openai",
		BaseURL: "https://api.openai.com",
		Models:  []string{"gpt-4"},
	}
	config, err := configService.CreateConfig(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Deactivate
	err = configService.DeactivateConfig(ctx, config.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify deactivated
	queriedConfig, err := configService.GetConfig(ctx, config.ID)
	if err != nil {
		t.Fatalf("Failed to get config: %v", err)
	}
	if queriedConfig.IsActive {
		t.Error("Expected config to be inactive")
	}

	// Activate
	err = configService.ActivateConfig(ctx, config.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify activated
	queriedConfig, err = configService.GetConfig(ctx, config.ID)
	if err != nil {
		t.Fatalf("Failed to get config: %v", err)
	}
	if !queriedConfig.IsActive {
		t.Error("Expected config to be active")
	}
}

// Unit test for GetConfigsByModel
func TestAPIConfigService_GetConfigsByModel(t *testing.T) {
	db := setupConfigTestDB(t)
	defer cleanupConfigTestDB(t, db)

	configRepo := repository.NewAPIConfigRepository(db)
	configService := NewAPIConfigService(configRepo)

	ctx := context.Background()

	// Create configs with different models
	req1 := &CreateConfigRequest{
		Name:    "Config 1",
		Type:    "openai",
		BaseURL: "https://api1.openai.com",
		Models:  []string{"gpt-4", "gpt-3.5-turbo"},
	}
	_, err := configService.CreateConfig(ctx, req1)
	if err != nil {
		t.Fatalf("Failed to create config 1: %v", err)
	}

	req2 := &CreateConfigRequest{
		Name:    "Config 2",
		Type:    "openai",
		BaseURL: "https://api2.openai.com",
		Models:  []string{"gpt-4"},
	}
	_, err = configService.CreateConfig(ctx, req2)
	if err != nil {
		t.Fatalf("Failed to create config 2: %v", err)
	}

	req3 := &CreateConfigRequest{
		Name:    "Config 3",
		Type:    "anthropic",
		BaseURL: "https://api.anthropic.com",
		Models:  []string{"claude-3"},
	}
	_, err = configService.CreateConfig(ctx, req3)
	if err != nil {
		t.Fatalf("Failed to create config 3: %v", err)
	}

	// Get configs for gpt-4
	configs, err := configService.GetConfigsByModel(ctx, "gpt-4")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(configs) != 2 {
		t.Errorf("Expected 2 configs for gpt-4, got %d", len(configs))
	}

	// Get configs for claude-3
	configs, err = configService.GetConfigsByModel(ctx, "claude-3")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(configs) != 1 {
		t.Errorf("Expected 1 config for claude-3, got %d", len(configs))
	}

	// Get configs for non-existent model
	configs, err = configService.GetConfigsByModel(ctx, "non-existent")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(configs) != 0 {
		t.Errorf("Expected 0 configs for non-existent model, got %d", len(configs))
	}
}
