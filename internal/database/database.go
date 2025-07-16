package database

import (
	"errors"
	"fmt"
	"log"
	"strings"

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
	if err := AddUniqueConstraints(db); err != nil {
		return nil, err
	}
	log.Println("Successfully connected to PostgreSQL database")
	return db, nil
}
func AddUniqueConstraints(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Проверяем существование таблиц
		if !tx.Migrator().HasTable("products") || !tx.Migrator().HasTable("nodes") {
			return errors.New("required tables don't exist")
		}

		// Добавляем ограничения с проверкой существования
		if err := addConstraintIfNotExists(tx,
			"products",
			"uc_products_pn_system",
			"ALTER TABLE products ADD CONSTRAINT uc_products_pn_system UNIQUE (pn, system_id)"); err != nil {
			return err
		}

		if err := addConstraintIfNotExists(tx,
			"nodes",
			"uc_nodes_name_system",
			"ALTER TABLE nodes ADD CONSTRAINT uc_nodes_name_system UNIQUE (name, system_id)"); err != nil {
			return err
		}

		return nil
	})
}
func addConstraintIfNotExists(db *gorm.DB, table, constraint, sql string) error {
	if !constraintExists(db, table, constraint) {
		if err := db.Exec(sql).Error; err != nil {
			// Игнорируем ошибку "уже существует" для разных СУБД
			if !strings.Contains(err.Error(), "already exists") &&
				!strings.Contains(err.Error(), "существует") {
				return fmt.Errorf("failed to add constraint %s: %w", constraint, err)
			}
		}
	}
	return nil
}
func constraintExists(db *gorm.DB, table, constraint string) bool {
	var count int
	db.Raw(`
		SELECT COUNT(*) 
		FROM information_schema.table_constraints 
		WHERE table_name = ? AND constraint_name = ?
	`, table, constraint).Scan(&count)
	return count > 0
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
