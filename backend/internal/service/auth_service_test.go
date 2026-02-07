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

// getTestDB creates a test database connection
func getTestDB(t *testing.T) *gorm.DB {
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

// setupTestDB initializes the test database
func setupTestDB(t *testing.T) *gorm.DB {
	db := getTestDB(t)

	// Clean up existing data
	db.Exec("TRUNCATE TABLE users CASCADE")

	// Run migrations
	err := db.AutoMigrate(&models.User{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

// cleanupTestDB cleans up the test database
func cleanupTestDB(t *testing.T, db *gorm.DB) {
	db.Exec("TRUNCATE TABLE users CASCADE")
}

// Property 1: 用户注册分配初始额度
// Feature: api-aggregator, Property 1: For any valid username, email, and password, when a user registers, the returned user object should have a quota of 10000 tokens.
// Validates: Requirements 1.1
func TestProperty_UserRegistrationAssignsInitialQuota(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret")

	properties := gopter.NewProperties(nil)

	// Custom generator for valid usernames (3-50 chars, alphanumeric)
	usernameGen := gen.RegexMatch("[a-zA-Z][a-zA-Z0-9_]{2,49}")
	
	// Custom generator for valid emails
	emailGen := gen.RegexMatch("[a-z]{3,20}@[a-z]{3,10}\\.com")
	
	// Custom generator for valid passwords (6-50 chars)
	passwordGen := gen.RegexMatch("[a-zA-Z0-9!@#$%]{6,50}")

	properties.Property("User registration assigns initial quota of 10000", prop.ForAll(
		func(username, email, password string) bool {
			// Clean up before each test
			db.Exec("TRUNCATE TABLE users CASCADE")

			ctx := context.Background()
			req := &RegisterRequest{
				Username: username,
				Email:    email,
				Password: password,
			}

			user, err := authService.Register(ctx, req)
			if err != nil {
				// If registration fails, it's not a property violation
				return true
			}

			// Check that the user has the initial quota of 10000
			return user.Quota == 10000
		},
		usernameGen,
		emailGen,
		passwordGen,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property 2: 登录返回有效JWT
// Feature: api-aggregator, Property 2: For any registered user, when logging in with correct credentials, the system should return a valid JWT token that can be decoded and contains the user_id.
// Validates: Requirements 1.2
func TestProperty_LoginReturnsValidJWT(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret")

	properties := gopter.NewProperties(nil)

	// Custom generators
	usernameGen := gen.RegexMatch("[a-zA-Z][a-zA-Z0-9_]{2,49}")
	emailGen := gen.RegexMatch("[a-z]{3,20}@[a-z]{3,10}\\.com")
	passwordGen := gen.RegexMatch("[a-zA-Z0-9!@#$%]{6,50}")

	properties.Property("Login returns valid JWT token", prop.ForAll(
		func(username, email, password string) bool {
			// Clean up before each test
			db.Exec("TRUNCATE TABLE users CASCADE")

			ctx := context.Background()

			// Register user first
			registerReq := &RegisterRequest{
				Username: username,
				Email:    email,
				Password: password,
			}
			user, err := authService.Register(ctx, registerReq)
			if err != nil {
				return true // Skip if registration fails
			}

			// Login with correct credentials
			loginReq := &LoginRequest{
				Username: username,
				Password: password,
			}
			authResp, err := authService.Login(ctx, loginReq)
			if err != nil {
				return false // Login should succeed
			}

			// Validate the token
			userID, err := authService.ValidateToken(authResp.Token)
			if err != nil {
				return false // Token should be valid
			}

			// Check that the user ID matches
			return userID == user.ID
		},
		usernameGen,
		emailGen,
		passwordGen,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property 3: 错误凭据拒绝登录
// Feature: api-aggregator, Property 3: For any invalid username or password combination, the login attempt should return a 401 Unauthorized error.
// Validates: Requirements 1.3
func TestProperty_InvalidCredentialsRejectLogin(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret")

	properties := gopter.NewProperties(nil)

	// Custom generators
	usernameGen := gen.RegexMatch("[a-zA-Z][a-zA-Z0-9_]{2,49}")
	emailGen := gen.RegexMatch("[a-z]{3,20}@[a-z]{3,10}\\.com")
	passwordGen := gen.RegexMatch("[a-zA-Z0-9!@#$%]{6,50}")
	wrongPasswordGen := gen.RegexMatch("[a-zA-Z0-9!@#$%]{6,50}")

	properties.Property("Invalid credentials reject login", prop.ForAll(
		func(username, email, password, wrongPassword string) bool {
			// Ensure passwords are different
			if password == wrongPassword {
				return true // Skip this case
			}

			// Clean up before each test
			db.Exec("TRUNCATE TABLE users CASCADE")

			ctx := context.Background()

			// Register user first
			registerReq := &RegisterRequest{
				Username: username,
				Email:    email,
				Password: password,
			}
			_, err := authService.Register(ctx, registerReq)
			if err != nil {
				return true // Skip if registration fails
			}

			// Try to login with wrong password
			loginReq := &LoginRequest{
				Username: username,
				Password: wrongPassword,
			}
			_, err = authService.Login(ctx, loginReq)

			// Should return ErrInvalidCredentials
			return err == ErrInvalidCredentials
		},
		usernameGen,
		emailGen,
		passwordGen,
		wrongPasswordGen,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property 4: 用户名和邮箱唯一性
// Feature: api-aggregator, Property 4: For any existing user, attempting to register with the same username or email should return a conflict error.
// Validates: Requirements 1.4
func TestProperty_UsernameAndEmailUniqueness(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret")

	properties := gopter.NewProperties(nil)

	// Custom generators
	usernameGen := gen.RegexMatch("[a-zA-Z][a-zA-Z0-9_]{2,39}") // Shorter to leave room for prefix
	emailGen := gen.RegexMatch("[a-z]{3,15}@[a-z]{3,10}\\.com") // Shorter to leave room for prefix
	passwordGen := gen.RegexMatch("[a-zA-Z0-9!@#$%]{6,50}")

	properties.Property("Username and email uniqueness", prop.ForAll(
		func(username, email, password string) bool {
			// Clean up before each test
			db.Exec("TRUNCATE TABLE users CASCADE")

			ctx := context.Background()

			// Register first user
			req1 := &RegisterRequest{
				Username: username,
				Email:    email,
				Password: password,
			}
			_, err := authService.Register(ctx, req1)
			if err != nil {
				return true // Skip if first registration fails
			}

			// Try to register with same username
			req2 := &RegisterRequest{
				Username: username,
				Email:    "different" + email,
				Password: password,
			}
			_, err = authService.Register(ctx, req2)
			if err != ErrUserExists {
				return false // Should return ErrUserExists
			}

			// Try to register with same email
			req3 := &RegisterRequest{
				Username: "different" + username,
				Email:    email,
				Password: password,
			}
			_, err = authService.Register(ctx, req3)
			if err != ErrUserExists {
				return false // Should return ErrUserExists
			}

			return true
		},
		usernameGen,
		emailGen,
		passwordGen,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Unit test to verify basic registration functionality
func TestAuthService_Register(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret")

	ctx := context.Background()

	// Test successful registration
	req := &RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	user, err := authService.Register(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if user.Username != req.Username {
		t.Errorf("Expected username %s, got %s", req.Username, user.Username)
	}
	if user.Email != req.Email {
		t.Errorf("Expected email %s, got %s", req.Email, user.Email)
	}
	if user.Quota != 10000 {
		t.Errorf("Expected quota 10000, got %d", user.Quota)
	}
	if user.UsedQuota != 0 {
		t.Errorf("Expected used quota 0, got %d", user.UsedQuota)
	}
	if user.IsAdmin {
		t.Error("Expected is_admin false, got true")
	}
	if user.Status != "active" {
		t.Errorf("Expected status 'active', got %s", user.Status)
	}

	// Test duplicate username
	req2 := &RegisterRequest{
		Username: "testuser",
		Email:    "test2@example.com",
		Password: "password123",
	}
	_, err = authService.Register(ctx, req2)
	if err != ErrUserExists {
		t.Errorf("Expected ErrUserExists, got %v", err)
	}

	// Test duplicate email
	req3 := &RegisterRequest{
		Username: "testuser2",
		Email:    "test@example.com",
		Password: "password123",
	}
	_, err = authService.Register(ctx, req3)
	if err != ErrUserExists {
		t.Errorf("Expected ErrUserExists, got %v", err)
	}
}

// Unit test to verify login functionality
func TestAuthService_Login(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret")

	ctx := context.Background()

	// Register a user first
	registerReq := &RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	user, err := authService.Register(ctx, registerReq)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Test successful login
	loginReq := &LoginRequest{
		Username: "testuser",
		Password: "password123",
	}
	authResp, err := authService.Login(ctx, loginReq)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if authResp.Token == "" {
		t.Error("Expected token, got empty string")
	}
	if authResp.User.ID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, authResp.User.ID)
	}

	// Test invalid username
	loginReq2 := &LoginRequest{
		Username: "wronguser",
		Password: "password123",
	}
	_, err = authService.Login(ctx, loginReq2)
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}

	// Test invalid password
	loginReq3 := &LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}
	_, err = authService.Login(ctx, loginReq3)
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

// Unit test to verify JWT token validation
func TestAuthService_ValidateToken(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret")

	ctx := context.Background()

	// Register and login a user
	registerReq := &RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	user, err := authService.Register(ctx, registerReq)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	loginReq := &LoginRequest{
		Username: "testuser",
		Password: "password123",
	}
	authResp, err := authService.Login(ctx, loginReq)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	// Test valid token
	userID, err := authService.ValidateToken(authResp.Token)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if userID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, userID)
	}

	// Test invalid token
	_, err = authService.ValidateToken("invalid-token")
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken, got %v", err)
	}

	// Test token with wrong secret
	wrongAuthService := NewAuthService(userRepo, "wrong-secret")
	_, err = wrongAuthService.ValidateToken(authResp.Token)
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken, got %v", err)
	}
}

