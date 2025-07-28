package main

import (
	"context"
	"fmt"
	"github.com/mejzh77/astragen/configs/config"
	"github.com/mejzh77/astragen/internal/api"
	"github.com/mejzh77/astragen/internal/database"
	"github.com/mejzh77/astragen/internal/sync"
	"gorm.io/gorm"
	"log"
)

func main() {
	// 1. Загрузка конфигурации
	log.Println("Loading configuration...")
	err := config.CreateDefaultConfigIfNotExist()
	if err != nil {
		log.Fatalf("Failed to create default config: %v", err)
	}
	config.Cfg = config.LoadConfig("config.yml")

	// 2. Инициализация БД
	log.Println("Initializing database connection...")
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s client_encoding=UTF8",
		config.Cfg.DB.Host,
		config.Cfg.DB.User,
		config.Cfg.DB.Password,
		config.Cfg.DB.Name,
		config.Cfg.DB.Port,
		config.Cfg.DB.SSLMode,
	)

	db, err := database.InitDB(dsn, true)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer closeDB(db)
	syncService := sync.NewSyncService(nil, db)
	ctx := context.Background()
	if config.Cfg.Update {
		err := syncService.RunFullSync(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}
	webService := api.NewWebService(syncService)

	// Добавляем функцию для шаблонов
	//webService.SetupTemplates()
	webService.RegisterRoutes()

	// Запуск сервера
	log.Println("Starting server on :8080")
	webService.Run(":8080")
}

func closeDB(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Failed to get SQL DB: %v", err)
		return
	}
	if err := sqlDB.Close(); err != nil {
		log.Printf("Failed to close database connection: %v", err)
	}
}
