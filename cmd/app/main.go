package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mejzh77/astragen/configs/config"
	"github.com/mejzh77/astragen/internal/database"
	"github.com/mejzh77/astragen/internal/gsheets"
)

func main() {
	// 1. Загрузка конфигурации
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s client_encoding=UTF8",
		config.AppConfig.DB.Host,
		config.AppConfig.DB.User,
		config.AppConfig.DB.Password,
		config.AppConfig.DB.Name,
		config.AppConfig.DB.Port,
		config.AppConfig.DB.SSLMode,
	)

	// 2. Инициализация БД
	db, err := database.InitDB(dsn)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()
	ctx := context.Background()
	creds, err := os.ReadFile("credentials.json")
	check(err)
	// Инициализация GORM
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Инициализация Google Sheets клиента
	sheetsService, err := gsheets.NewService(ctx, creds)
	if err != nil {
		log.Fatal("Failed to create Google Sheets service:", err)
	}

	// Создание нашего сервиса
	gsheetsService := gsheets.NewGoogleSheetsService(sheetsService, db)
	check(err)
	err = gsheetsService.RunSync(ctx)
	check(err)
}

func check(err error) {
	if err != nil {
		log.Fatalf("%v", err)
	}
}
