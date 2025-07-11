package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mejzh77/astragen/configs/config"
	"github.com/mejzh77/astragen/internal/database"
	"github.com/mejzh77/astragen/internal/gsheets"
	"github.com/mejzh77/astragen/internal/sync"
	"github.com/mejzh77/astragen/pkg/models"
	"gorm.io/gorm"
)

func main() {
	// 1. Загрузка конфигурации
	log.Println("Loading configuration...")
	config.Cfg = config.LoadConfig("configs/config.yml")

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

	db, err := database.InitDB(dsn)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer closeDB(db)

	if !config.Cfg.Update {
		log.Println("Update flag is false, exiting")
		return
	}

	// 3. Инициализация сервисов
	log.Println("Initializing services...")
	ctx := context.Background()
	creds, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Failed to read credentials file: %v", err)
	}

	sheetsService, err := gsheets.NewService(ctx, creds)
	if err != nil {
		log.Fatalf("Failed to create Google Sheets service: %v", err)
	}

	syncService := sync.NewSyncService(sheetsService, db)

	// 4. Полная синхронизация с логированием
	log.Println("Starting full sync process...")

	// 4.2. Синхронизация сигналов
	log.Println("Syncing signals...")
	signals, err := syncService.LoadAndSaveSignals(ctx)
	if err != nil {
		log.Fatalf("Failed to sync signals: %v", err)
	}
	logSignals(db)

	// 4.1. Синхронизация проектов и систем
	log.Println("Syncing projects and systems...")
	if err := syncService.SyncSystemsFromSignals(signals); err != nil {
		log.Fatalf("Failed to sync projects and systems: %v", err)
	}
	logProjectsAndSystems(db)
	// 4.3. Синхронизация узлов (поддержка нескольких систем)
	log.Println("Syncing nodes...")
	if err := syncService.SyncNodes(signals); err != nil {
		log.Fatalf("Failed to sync nodes: %v", err)
	}
	logNodes(db)

	// 4.4. Синхронизация продуктов
	log.Println("Syncing products...")
	if err := syncService.SyncProducts(signals); err != nil {
		log.Fatalf("Failed to sync products: %v", err)
	}
	logProducts(db)

	// 4.5. Синхронизация функциональных блоков
	log.Println("Syncing function blocks...")
	if err := syncService.SyncFunctionBlocks(signals); err != nil {
		log.Fatalf("Failed to sync function blocks: %v", err)
	}
	logFunctionBlocks(db)

	log.Println("Sync completed successfully!")
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

// Функции логирования данных
func logProjectsAndSystems(db *gorm.DB) {
	var projects []models.Project
	db.Preload("Systems").Find(&projects)
	log.Printf("Imported projects and systems: %+v", projects)
}

func logSignals(db *gorm.DB) {
	var count int64
	db.Model(&models.Signal{}).Count(&count)
	log.Printf("Imported signals count: %d", count)
}

func logNodes(db *gorm.DB) {
	var nodes []models.Node
	db.Preload("Systems").Find(&nodes)
	log.Printf("Imported nodes: %+v", nodes)
}

func logProducts(db *gorm.DB) {
	var products []models.Product
	db.Preload("System").Find(&products)
	log.Printf("Imported products: %+v", products)
}

func logFunctionBlocks(db *gorm.DB) {
	var fbs []models.FunctionBlock
	db.Preload("Variables").Find(&fbs)
	log.Printf("Imported function blocks: %+v", fbs)
}
