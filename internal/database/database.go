package database

import (
	"fmt"
	"log"

	"github.com/mejzh77/astragen/pkg/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// InitDB инициализирует и возвращает подключение к PostgreSQL
func InitDB(dsn string) (*gorm.DB, error) {
	// Подключение к БД
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// Дополнительные настройки GORM
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Проверка соединения
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Настройка пула соединений
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	// Автомиграция моделей
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate models: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database")
	return db, nil
}

func autoMigrate(db *gorm.DB) error {
	modelsToMigrate := []interface{}{
		&models.Signal{},
		&models.Product{},
		&models.Node{},
		&models.FunctionBlock{},
		&models.FBVariable{},
		&models.Project{},
		// Добавьте другие модели по мере необходимости
	}

	for _, model := range modelsToMigrate {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to auto-migrate model %T: %w", model, err)
		}
	}
	return nil
}
