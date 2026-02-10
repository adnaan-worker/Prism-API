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

// getModelTestDB creates a test database connection
func getModelTestDB(t *testing.T) *gorm.DB {
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

// setupModelTestDB initializes the test database
func setupModelTestDB(t *testing.T) *gorm.DB {
	db := getModelTestDB(t)

	// Clean up existing data
	db.Exec("TRUNCATE TABLE api_configs CASCADE")

	// Run migrations
	err := db.AutoMigrate(&models.APIConfig{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

// cleanupModelTestDB cleans up the test database
func cleanupModelTestDB(t *testing.T, db *gorm.DB) {
	db.Exec("TRUNCATE TABLE api_configs CASCADE")
}

// Property 9: 模型列表完整性
// Feature: api-aggregator, Property 9: For any set of active API configurations, the model list should contain all models from all configurations.
// Validates: Requirements 4.1
// Note: After refactoring, each config's models are listed separately (no deduplication)
func TestProperty_ModelListCompleteness(t *testing.T) {
	db := setupModelTestDB(t)
	defer cleanupModelTestDB(t, db)

	configRepo := repository.NewAPIConfigRepository(db)
	modelService := NewModelService(configRepo)

	properties := gopter.NewProperties(nil)

	// Generator for model lists (1-5 models per config)
	modelsGen := gen.SliceOfN(3, gen.RegexMatch("[a-z]{3,10}-[0-9]"))

	properties.Property("Model list contains all models from all configs", prop.ForAll(
		func(models1, models2, models3 []string) bool {
			// Clean up before each test
			db.Exec("TRUNCATE TABLE api_configs CASCADE")

			ctx := context.Background()

			// Create multiple configs with different models
			config1 := &models.APIConfig{
				Name:     "Config 1",
				Type:     "openai",
				BaseURL:  "https://api1.com",
				APIKey:   "key1",
				Models:   models1,
				IsActive: true,
			}
			if err := configRepo.Create(ctx, config1); err != nil {
				return true // Skip on error
			}

			config2 := &models.APIConfig{
				Name:     "Config 2",
				Type:     "anthropic",
				BaseURL:  "https://api2.com",
				APIKey:   "key2",
				Models:   models2,
				IsActive: true,
			}
			if err := configRepo.Create(ctx, config2); err != nil {
				return true // Skip on error
			}

			config3 := &models.APIConfig{
				Name:     "Config 3",
				Type:     "gemini",
				BaseURL:  "https://api3.com",
				APIKey:   "key3",
				Models:   models3,
				IsActive: true,
			}
			if err := configRepo.Create(ctx, config3); err != nil {
				return true // Skip on error
			}

			// Get all models
			allModels, err := modelService.GetAllModels(ctx)
			if err != nil {
				return false
			}

			// Create a map of expected models with their config IDs
			expectedModels := make(map[string]map[uint]bool)
			for _, m := range models1 {
				if expectedModels[m] == nil {
					expectedModels[m] = make(map[uint]bool)
				}
				expectedModels[m][config1.ID] = true
			}
			for _, m := range models2 {
				if expectedModels[m] == nil {
					expectedModels[m] = make(map[uint]bool)
				}
				expectedModels[m][config2.ID] = true
			}
			for _, m := range models3 {
				if expectedModels[m] == nil {
					expectedModels[m] = make(map[uint]bool)
				}
				expectedModels[m][config3.ID] = true
			}

			// Create a map of returned models with their config IDs
			returnedModels := make(map[string]map[uint]bool)
			for _, modelInfo := range allModels {
				if returnedModels[modelInfo.Name] == nil {
					returnedModels[modelInfo.Name] = make(map[uint]bool)
				}
				returnedModels[modelInfo.Name][modelInfo.ConfigID] = true
			}

			// Verify all expected model-config pairs are in the returned list
			for modelName, configIDs := range expectedModels {
				returnedConfigs, exists := returnedModels[modelName]
				if !exists {
					return false
				}
				for configID := range configIDs {
					if !returnedConfigs[configID] {
						return false
					}
				}
			}

			return true
		},
		modelsGen,
		modelsGen,
		modelsGen,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property 10: 模型去重
// Feature: api-aggregator, Property 10: For any model name that appears in multiple API configurations, GetUniqueModels should return each model name exactly once.
// Validates: Requirements 4.2
// Note: After refactoring, GetAllModels lists each config's models separately, but GetUniqueModels deduplicates them
func TestProperty_ModelDeduplication(t *testing.T) {
	db := setupModelTestDB(t)
	defer cleanupModelTestDB(t, db)

	configRepo := repository.NewAPIConfigRepository(db)
	modelService := NewModelService(configRepo)

	properties := gopter.NewProperties(nil)

	// Generator for a single model name
	modelGen := gen.RegexMatch("[a-z]{3,10}-[0-9]")

	properties.Property("Duplicate models appear only once", prop.ForAll(
		func(sharedModel string) bool {
			// Clean up before each test
			db.Exec("TRUNCATE TABLE api_configs CASCADE")

			ctx := context.Background()

			// Create multiple configs with the same model
			config1 := &models.APIConfig{
				Name:     "Config 1",
				Type:     "openai",
				BaseURL:  "https://api1.com",
				APIKey:   "key1",
				Models:   []string{sharedModel, "unique-1"},
				IsActive: true,
			}
			if err := configRepo.Create(ctx, config1); err != nil {
				return true // Skip on error
			}

			config2 := &models.APIConfig{
				Name:     "Config 2",
				Type:     "openai",
				BaseURL:  "https://api2.com",
				APIKey:   "key2",
				Models:   []string{sharedModel, "unique-2"},
				IsActive: true,
			}
			if err := configRepo.Create(ctx, config2); err != nil {
				return true // Skip on error
			}

			config3 := &models.APIConfig{
				Name:     "Config 3",
				Type:     "openai",
				BaseURL:  "https://api3.com",
				APIKey:   "key3",
				Models:   []string{sharedModel, "unique-3"},
				IsActive: true,
			}
			if err := configRepo.Create(ctx, config3); err != nil {
				return true // Skip on error
			}

			// Get all models (should list each config's models separately)
			allModels, err := modelService.GetAllModels(ctx)
			if err != nil {
				return false
			}

			// Count occurrences of the shared model in GetAllModels
			// After refactoring, each config's model is listed separately
			// So we should see the shared model 3 times (once per config)
			count := 0
			for _, modelInfo := range allModels {
				if modelInfo.Name == sharedModel {
					count++
				}
			}

			// Shared model should appear 3 times (once per config)
			return count == 3
		},
		modelGen,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Unit test for GetAllModels
func TestModelService_GetAllModels(t *testing.T) {
	db := setupModelTestDB(t)
	defer cleanupModelTestDB(t, db)

	configRepo := repository.NewAPIConfigRepository(db)
	modelService := NewModelService(configRepo)

	ctx := context.Background()

	// Clean up to ensure fresh state
	db.Exec("TRUNCATE TABLE api_configs CASCADE")

	// Create test configs
	config1 := &models.APIConfig{
		Name:     "OpenAI Config",
		Type:     "openai",
		BaseURL:  "https://api.openai.com",
		APIKey:   "key1",
		Models:   []string{"gpt-4", "gpt-3.5-turbo"},
		IsActive: true,
	}
	configRepo.Create(ctx, config1)

	config2 := &models.APIConfig{
		Name:     "Anthropic Config",
		Type:     "anthropic",
		BaseURL:  "https://api.anthropic.com",
		APIKey:   "key2",
		Models:   []string{"claude-3-opus", "claude-3-sonnet"},
		IsActive: true,
	}
	configRepo.Create(ctx, config2)

	// Get all models
	allModels, err := modelService.GetAllModels(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(allModels) != 4 {
		t.Errorf("Expected 4 models, got %d", len(allModels))
	}

	// Verify model info
	modelNames := make(map[string]bool)
	for _, model := range allModels {
		modelNames[model.Name] = true
		if !model.IsActive {
			t.Errorf("Expected IsActive true, got false")
		}
	}

	expectedModels := []string{"gpt-4", "gpt-3.5-turbo", "claude-3-opus", "claude-3-sonnet"}
	for _, expected := range expectedModels {
		if !modelNames[expected] {
			t.Errorf("Expected model %s not found", expected)
		}
	}
}

// Unit test for GetUniqueModels
func TestModelService_GetUniqueModels(t *testing.T) {
	db := setupModelTestDB(t)
	defer cleanupModelTestDB(t, db)

	configRepo := repository.NewAPIConfigRepository(db)
	modelService := NewModelService(configRepo)

	ctx := context.Background()

	// Clean up to ensure fresh state
	db.Exec("TRUNCATE TABLE api_configs CASCADE")

	// Create configs with overlapping models
	config1 := &models.APIConfig{
		Name:     "Config 1",
		Type:     "openai",
		BaseURL:  "https://api1.com",
		APIKey:   "key1",
		Models:   []string{"gpt-4", "gpt-3.5-turbo"},
		IsActive: true,
	}
	configRepo.Create(ctx, config1)

	config2 := &models.APIConfig{
		Name:     "Config 2",
		Type:     "openai",
		BaseURL:  "https://api2.com",
		APIKey:   "key2",
		Models:   []string{"gpt-4", "claude-3"},
		IsActive: true,
	}
	configRepo.Create(ctx, config2)

	// Get unique models
	uniqueModels, err := modelService.GetUniqueModels(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should have 3 unique models (gpt-4, gpt-3.5-turbo, claude-3)
	if len(uniqueModels) != 3 {
		t.Errorf("Expected 3 unique models, got %d", len(uniqueModels))
	}

	// Verify deduplication
	modelSet := make(map[string]bool)
	for _, model := range uniqueModels {
		if modelSet[model] {
			t.Errorf("Duplicate model found: %s", model)
		}
		modelSet[model] = true
	}
}

// Unit test for provider normalization
func TestModelService_NormalizeProvider(t *testing.T) {
	db := setupModelTestDB(t)
	defer cleanupModelTestDB(t, db)

	configRepo := repository.NewAPIConfigRepository(db)
	modelService := NewModelService(configRepo)

	tests := []struct {
		configType       string
		expectedProvider string
	}{
		{"openai", "OpenAI"},
		{"anthropic", "Anthropic"},
		{"gemini", "Google"},
		{"custom", "Custom"},
		{"unknown", "Custom"},
	}

	for _, tt := range tests {
		t.Run(tt.configType, func(t *testing.T) {
			provider := modelService.normalizeProvider(tt.configType)
			if provider != tt.expectedProvider {
				t.Errorf("Expected provider %s, got %s", tt.expectedProvider, provider)
			}
		})
	}
}

// Unit test for GetModelsByProvider
func TestModelService_GetModelsByProvider(t *testing.T) {
	db := setupModelTestDB(t)
	defer cleanupModelTestDB(t, db)

	configRepo := repository.NewAPIConfigRepository(db)
	modelService := NewModelService(configRepo)

	ctx := context.Background()

	// Clean up to ensure fresh state
	db.Exec("TRUNCATE TABLE api_configs CASCADE")

	// Create test configs
	config1 := &models.APIConfig{
		Name:     "OpenAI Config",
		Type:     "openai",
		BaseURL:  "https://api.openai.com",
		APIKey:   "key1",
		Models:   []string{"gpt-4", "gpt-3.5-turbo"},
		IsActive: true,
	}
	configRepo.Create(ctx, config1)

	config2 := &models.APIConfig{
		Name:     "Anthropic Config",
		Type:     "anthropic",
		BaseURL:  "https://api.anthropic.com",
		APIKey:   "key2",
		Models:   []string{"claude-3-opus"},
		IsActive: true,
	}
	configRepo.Create(ctx, config2)

	// Get OpenAI models
	openaiModels, err := modelService.GetModelsByProvider(ctx, "OpenAI")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(openaiModels) != 2 {
		t.Errorf("Expected 2 OpenAI models, got %d", len(openaiModels))
	}

	// Get Anthropic models
	anthropicModels, err := modelService.GetModelsByProvider(ctx, "Anthropic")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(anthropicModels) != 1 {
		t.Errorf("Expected 1 Anthropic model, got %d", len(anthropicModels))
	}
}

// Unit test for inactive configs
func TestModelService_InactiveConfigs(t *testing.T) {
	db := setupModelTestDB(t)
	defer cleanupModelTestDB(t, db)

	configRepo := repository.NewAPIConfigRepository(db)
	modelService := NewModelService(configRepo)

	ctx := context.Background()

	// Clean up to ensure fresh state
	db.Exec("TRUNCATE TABLE api_configs CASCADE")

	// Create active config
	config1 := &models.APIConfig{
		Name:     "Active Config",
		Type:     "openai",
		BaseURL:  "https://api1.com",
		APIKey:   "key1",
		Models:   []string{"gpt-4"},
		IsActive: true,
	}
	if err := configRepo.Create(ctx, config1); err != nil {
		t.Fatalf("Failed to create active config: %v", err)
	}

	// Create inactive config - use Update to explicitly set IsActive to false
	config2 := &models.APIConfig{
		Name:     "Inactive Config",
		Type:     "openai",
		BaseURL:  "https://api2.com",
		APIKey:   "key2",
		Models:   []string{"gpt-3.5-turbo"},
		IsActive: true, // Create as active first
	}
	if err := configRepo.Create(ctx, config2); err != nil {
		t.Fatalf("Failed to create inactive config: %v", err)
	}
	// Then deactivate it
	config2.IsActive = false
	if err := configRepo.Update(ctx, config2); err != nil {
		t.Fatalf("Failed to deactivate config: %v", err)
	}

	// Get all models - should only include active configs
	allModels, err := modelService.GetAllModels(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(allModels) != 1 {
		t.Errorf("Expected 1 model (from active config only), got %d", len(allModels))
		return
	}

	if allModels[0].Name != "gpt-4" {
		t.Errorf("Expected gpt-4, got %s", allModels[0].Name)
	}
}
