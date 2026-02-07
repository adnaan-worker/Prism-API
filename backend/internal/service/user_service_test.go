package service

import (
	"api-aggregator/backend/internal/models"
	"api-aggregator/backend/internal/repository"
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupUserServiceTest(t *testing.T) (*UserService, *gorm.DB) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "host=localhost user=postgres password=postgres dbname=api_aggregator_test port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Clean up existing data
	db.Exec("TRUNCATE TABLE users CASCADE")

	// Run migrations
	err = db.AutoMigrate(&models.User{})
	require.NoError(t, err)

	userRepo := repository.NewUserRepository(db)
	userService := NewUserService(userRepo)

	return userService, db
}

func createTestUser(t *testing.T, db *gorm.DB, username, email string, isAdmin bool) *models.User {
	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: "hashed_password",
		Quota:        10000,
		UsedQuota:    0,
		IsAdmin:      isAdmin,
		Status:       "active",
	}
	err := db.Create(user).Error
	require.NoError(t, err)
	return user
}

func TestGetUsers(t *testing.T) {
	userService, db := setupUserServiceTest(t)
	ctx := context.Background()

	// Create test users
	createTestUser(t, db, "user1", "user1@example.com", false)
	createTestUser(t, db, "user2", "user2@example.com", false)
	createTestUser(t, db, "admin", "admin@example.com", true)

	t.Run("get users with default pagination", func(t *testing.T) {
		req := &GetUsersRequest{}
		resp, err := userService.GetUsers(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, int64(3), resp.Total)
		assert.Equal(t, 1, resp.Page)
		assert.Equal(t, 10, resp.PageSize)
		assert.Len(t, resp.Users, 3)
	})

	t.Run("get users with custom pagination", func(t *testing.T) {
		req := &GetUsersRequest{
			Page:     1,
			PageSize: 2,
		}
		resp, err := userService.GetUsers(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, int64(3), resp.Total)
		assert.Equal(t, 1, resp.Page)
		assert.Equal(t, 2, resp.PageSize)
		assert.Len(t, resp.Users, 2)
	})

	t.Run("get second page", func(t *testing.T) {
		req := &GetUsersRequest{
			Page:     2,
			PageSize: 2,
		}
		resp, err := userService.GetUsers(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, int64(3), resp.Total)
		assert.Equal(t, 2, resp.Page)
		assert.Len(t, resp.Users, 1)
	})

	t.Run("invalid page size", func(t *testing.T) {
		req := &GetUsersRequest{
			Page:     1,
			PageSize: 101,
		}
		_, err := userService.GetUsers(ctx, req)

		assert.ErrorIs(t, err, ErrInvalidPage)
	})
}

func TestGetUserByID(t *testing.T) {
	userService, db := setupUserServiceTest(t)
	ctx := context.Background()

	user := createTestUser(t, db, "testuser", "test@example.com", false)

	t.Run("get existing user", func(t *testing.T) {
		result, err := userService.GetUserByID(ctx, user.ID)

		require.NoError(t, err)
		assert.Equal(t, user.ID, result.ID)
		assert.Equal(t, user.Username, result.Username)
		assert.Equal(t, user.Email, result.Email)
	})

	t.Run("get non-existent user", func(t *testing.T) {
		_, err := userService.GetUserByID(ctx, 99999)

		assert.ErrorIs(t, err, ErrUserNotFound)
	})
}

func TestUpdateUserStatus(t *testing.T) {
	userService, db := setupUserServiceTest(t)
	ctx := context.Background()

	user := createTestUser(t, db, "testuser", "test@example.com", false)

	t.Run("update status successfully", func(t *testing.T) {
		err := userService.UpdateUserStatus(ctx, user.ID, "inactive")
		require.NoError(t, err)

		// Verify status was updated
		var updatedUser models.User
		err = db.First(&updatedUser, user.ID).Error
		require.NoError(t, err)
		assert.Equal(t, "inactive", updatedUser.Status)
	})

	t.Run("update status of non-existent user", func(t *testing.T) {
		err := userService.UpdateUserStatus(ctx, 99999, "inactive")
		assert.ErrorIs(t, err, ErrUserNotFound)
	})
}

func TestUpdateUserQuota(t *testing.T) {
	userService, db := setupUserServiceTest(t)
	ctx := context.Background()

	user := createTestUser(t, db, "testuser", "test@example.com", false)

	t.Run("update quota successfully", func(t *testing.T) {
		newQuota := int64(50000)
		err := userService.UpdateUserQuota(ctx, user.ID, newQuota)
		require.NoError(t, err)

		// Verify quota was updated
		var updatedUser models.User
		err = db.First(&updatedUser, user.ID).Error
		require.NoError(t, err)
		assert.Equal(t, newQuota, updatedUser.Quota)
	})

	t.Run("update quota of non-existent user", func(t *testing.T) {
		err := userService.UpdateUserQuota(ctx, 99999, 50000)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})
}
