package main

import (
	"api-aggregator/backend/config"
	"api-aggregator/backend/internal/models"
	"context"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Get admin credentials from environment or use defaults
	adminUsername := os.Getenv("ADMIN_USERNAME")
	if adminUsername == "" {
		adminUsername = cfg.Admin.Username
		if adminUsername == "" {
			adminUsername = "admin"
		}
	}

	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = cfg.Admin.Email
		if adminEmail == "" {
			adminEmail = "admin@example.com"
		}
	}

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = cfg.Admin.Password
		if adminPassword == "" {
			adminPassword = "admin123"
		}
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.Database.URL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Check if users table exists (run migrations first if needed)
	var tableExists bool
	db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'users')").Scan(&tableExists)
	if !tableExists {
		log.Println("Users table does not exist. Running migrations first...")
		if err := db.AutoMigrate(&models.User{}); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		log.Println("Migrations completed successfully.")
	}

	ctx := context.Background()

	// Check if admin user already exists
	var existingUser models.User
	result := db.WithContext(ctx).Where("username = ?", adminUsername).First(&existingUser)
	if result.Error == nil {
		log.Printf("Admin user '%s' already exists (ID: %d, Email: %s)", 
			existingUser.Username, existingUser.ID, existingUser.Email)
		
		// Check if --force flag is set for non-interactive mode
		forceUpdate := false
		for _, arg := range os.Args[1:] {
			if arg == "--force" || arg == "-f" {
				forceUpdate = true
				break
			}
		}
		
		if !forceUpdate {
			// Ask if user wants to update password (only in interactive mode)
			fmt.Print("Do you want to update the password? (y/n): ")
			var answer string
			fmt.Scanln(&answer)
			forceUpdate = (answer == "y" || answer == "Y")
		}
		
		if forceUpdate {
			// Hash new password
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
			if err != nil {
				log.Fatalf("Failed to hash password: %v", err)
			}
			
			// Update password and ensure admin status
			existingUser.PasswordHash = string(hashedPassword)
			existingUser.IsAdmin = true
			existingUser.Status = "active"
			
			if err := db.WithContext(ctx).Save(&existingUser).Error; err != nil {
				log.Fatalf("Failed to update admin user: %v", err)
			}
			
			log.Printf("Successfully updated admin user '%s' password", adminUsername)
		} else {
			log.Println("Password update skipped.")
		}
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Create admin user
	adminUser := &models.User{
		Username:     adminUsername,
		Email:        adminEmail,
		PasswordHash: string(hashedPassword),
		Quota:        100000, // Higher quota for admin
		UsedQuota:    0,
		IsAdmin:      true,
		Status:       "active",
	}

	if err := db.WithContext(ctx).Create(adminUser).Error; err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	log.Printf("Successfully created admin user:")
	log.Printf("  Username: %s", adminUsername)
	log.Printf("  Email: %s", adminEmail)
	log.Printf("  Password: [HIDDEN - check your .env file or environment variables]")
	log.Printf("  ID: %d", adminUser.ID)
	log.Printf("\nYou can now login with these credentials.")
}
