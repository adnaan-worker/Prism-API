package main

import (
	"api-aggregator/backend/config"
	"api-aggregator/backend/internal/app"
	"fmt"
	"log"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// 创建应用实例
	application, err := app.New(cfg)
	if err != nil {
		log.Fatal("Failed to initialize application:", err)
	}
	defer application.Close()

	// 启动服务器
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	if err := application.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
