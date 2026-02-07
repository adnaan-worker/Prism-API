package middleware

import (
	"api-aggregator/backend/internal/models"
	"api-aggregator/backend/internal/repository"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupAdminTestDB(t *testing.T) *gorm.DB {
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

	// Clean up existing data
	db.Exec("TRUNCATE TABLE users CASCADE")

	// Migrate the schema
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func cleanupAdminTestDB(t *testing.T, db *gorm.DB) {
	db.Exec("TRUNCATE TABLE users CASCADE")
}

func TestAdminMiddleware_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAdminTestDB(t)
	defer cleanupAdminTestDB(t, db)
	userRepo := repository.NewUserRepository(db)

	// Create an admin user
	adminUser := &models.User{
		Username:     "admin",
		Email:        "admin@example.com",
		PasswordHash: "hash",
		IsAdmin:      true,
		Quota:        10000,
		UsedQuota:    0,
		Status:       "active",
	}
	err := userRepo.Create(context.Background(), adminUser)
	assert.NoError(t, err)

	// Setup router
	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Simulate AuthMiddleware setting user_id
		c.Set("user_id", adminUser.ID)
		c.Next()
	})
	router.Use(AdminMiddleware(userRepo))
	router.GET("/admin/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin/test", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["message"])
}

func TestAdminMiddleware_NonAdminUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAdminTestDB(t)
	defer cleanupAdminTestDB(t, db)
	userRepo := repository.NewUserRepository(db)

	// Create a regular user (not admin)
	regularUser := &models.User{
		Username:     "user",
		Email:        "user@example.com",
		PasswordHash: "hash",
		IsAdmin:      false,
		Quota:        10000,
		UsedQuota:    0,
		Status:       "active",
	}
	err := userRepo.Create(context.Background(), regularUser)
	assert.NoError(t, err)

	// Setup router
	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Simulate AuthMiddleware setting user_id
		c.Set("user_id", regularUser.ID)
		c.Next()
	})
	router.Use(AdminMiddleware(userRepo))
	router.GET("/admin/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin/test", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	errorMap := response["error"].(map[string]interface{})
	assert.Equal(t, float64(403001), errorMap["code"])
	assert.Equal(t, "Forbidden", errorMap["message"])
	assert.Equal(t, "Admin privileges required", errorMap["details"])
}

func TestAdminMiddleware_NoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAdminTestDB(t)
	defer cleanupAdminTestDB(t, db)
	userRepo := repository.NewUserRepository(db)

	// Setup router without setting user_id
	router := gin.New()
	router.Use(AdminMiddleware(userRepo))
	router.GET("/admin/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin/test", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	errorMap := response["error"].(map[string]interface{})
	assert.Equal(t, float64(401001), errorMap["code"])
	assert.Equal(t, "Unauthorized", errorMap["message"])
	assert.Equal(t, "User not authenticated", errorMap["details"])
}

func TestAdminMiddleware_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAdminTestDB(t)
	defer cleanupAdminTestDB(t, db)
	userRepo := repository.NewUserRepository(db)

	// Setup router with non-existent user ID
	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Set a user_id that doesn't exist
		c.Set("user_id", uint(999))
		c.Next()
	})
	router.Use(AdminMiddleware(userRepo))
	router.GET("/admin/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin/test", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	errorMap := response["error"].(map[string]interface{})
	assert.Equal(t, float64(401001), errorMap["code"])
	assert.Equal(t, "Unauthorized", errorMap["message"])
	assert.Equal(t, "User not found", errorMap["details"])
}

func TestAdminMiddleware_SetsUserInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAdminTestDB(t)
	defer cleanupAdminTestDB(t, db)
	userRepo := repository.NewUserRepository(db)

	// Create an admin user
	adminUser := &models.User{
		Username:     "admin",
		Email:        "admin@example.com",
		PasswordHash: "hash",
		IsAdmin:      true,
		Quota:        10000,
		UsedQuota:    0,
		Status:       "active",
	}
	err := userRepo.Create(context.Background(), adminUser)
	assert.NoError(t, err)

	// Setup router
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", adminUser.ID)
		c.Next()
	})
	router.Use(AdminMiddleware(userRepo))
	router.GET("/admin/test", func(c *gin.Context) {
		// Check if user is set in context
		userValue, exists := c.Get("user")
		assert.True(t, exists)
		user, ok := userValue.(*models.User)
		assert.True(t, ok)
		assert.Equal(t, adminUser.ID, user.ID)
		assert.Equal(t, adminUser.Username, user.Username)
		assert.True(t, user.IsAdmin)
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin/test", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminMiddleware_InvalidUserIDType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAdminTestDB(t)
	defer cleanupAdminTestDB(t, db)
	userRepo := repository.NewUserRepository(db)

	// Setup router with invalid user_id type
	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Set user_id as string instead of uint
		c.Set("user_id", "invalid")
		c.Next()
	})
	router.Use(AdminMiddleware(userRepo))
	router.GET("/admin/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin/test", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	errorMap := response["error"].(map[string]interface{})
	assert.Equal(t, float64(500001), errorMap["code"])
	assert.Equal(t, "Internal Error", errorMap["message"])
	assert.Equal(t, "Invalid user ID format", errorMap["details"])
}

func TestAdminMiddleware_DeletedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAdminTestDB(t)
	defer cleanupAdminTestDB(t, db)
	userRepo := repository.NewUserRepository(db)

	// Create an admin user
	adminUser := &models.User{
		Username:     "admin",
		Email:        "admin@example.com",
		PasswordHash: "hash",
		IsAdmin:      true,
		Quota:        10000,
		UsedQuota:    0,
		Status:       "active",
	}
	err := userRepo.Create(context.Background(), adminUser)
	assert.NoError(t, err)

	// Soft delete the user
	now := time.Now()
	adminUser.DeletedAt = gorm.DeletedAt{Time: now, Valid: true}
	err = db.Save(adminUser).Error
	assert.NoError(t, err)

	// Setup router
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", adminUser.ID)
		c.Next()
	})
	router.Use(AdminMiddleware(userRepo))
	router.GET("/admin/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin/test", nil)
	router.ServeHTTP(w, req)

	// Assert - deleted user should not be found
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	errorMap := response["error"].(map[string]interface{})
	assert.Equal(t, float64(401001), errorMap["code"])
	assert.Equal(t, "Unauthorized", errorMap["message"])
	assert.Equal(t, "User not found", errorMap["details"])
}
