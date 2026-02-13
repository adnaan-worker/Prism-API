package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type AccountCredential struct {
	ID            uint   `gorm:"primarykey"`
	ProviderType  string `gorm:"column:provider_type"`
	HealthStatus  string
	TotalRequests int64
	TotalErrors   int64
	IsActive      bool `gorm:"column:is_active"`
}

func (AccountCredential) TableName() string {
	return "account_credentials"
}

func main() {
	// 从环境变量获取数据库连接
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/api_aggregator?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 查询所有凭据
	var creds []AccountCredential
	if err := db.Find(&creds).Error; err != nil {
		log.Fatal("Failed to query credentials:", err)
	}

	fmt.Println("=== 当前凭据状态 ===")
	fmt.Printf("%-5s %-15s %-15s %-15s %-15s %-10s\n", "ID", "Provider", "HealthStatus", "TotalRequests", "TotalErrors", "IsActive")
	fmt.Println("------------------------------------------------------------------------------------")
	
	for _, cred := range creds {
		fmt.Printf("%-5d %-15s %-15s %-15d %-15d %-10t\n",
			cred.ID,
			cred.ProviderType,
			cred.HealthStatus,
			cred.TotalRequests,
			cred.TotalErrors,
			cred.IsActive,
		)
	}

	// 询问是否重置
	fmt.Println("\n是否重置所有凭据的健康状态为 'healthy' 并清零错误计数? (y/n)")
	var answer string
	fmt.Scanln(&answer)

	if answer != "y" && answer != "Y" {
		fmt.Println("取消操作")
		return
	}

	// 重置健康状态
	result := db.Model(&AccountCredential{}).
		Where("1=1").
		Updates(map[string]interface{}{
			"health_status": "healthy",
			"total_errors":  0,
		})

	if result.Error != nil {
		log.Fatal("Failed to reset health status:", result.Error)
	}

	fmt.Printf("\n成功重置 %d 个凭据的健康状态\n", result.RowsAffected)
}
