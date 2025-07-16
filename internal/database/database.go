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
func InitDB(dsn string, cleanBeforeMigrate bool) (*gorm.DB, error) {
	// Подключение к БД
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
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

	// Очистка таблиц перед миграцией (если требуется)
	if cleanBeforeMigrate {
		if err := cleanDatabase(db); err != nil {
			return nil, fmt.Errorf("failed to clean database: %w", err)
		}
	}

	// Автомиграция моделей
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate models: %w", err)
	}

	if err := AddUniqueConstraints(db); err != nil {
		return nil, err
	}
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm").Error; err != nil {
		return nil, fmt.Errorf("failed to create pg_trgm extension: %w", err)
	}
	log.Println("Successfully connected to PostgreSQL database")
	return db, nil
}

// cleanDatabase полностью очищает все таблицы в правильном порядке
func cleanDatabase(db *gorm.DB) error {
	tables := []string{
		"fb_variables",    // Зависит от function_blocks
		"signals",         // Зависит от nodes и products
		"function_blocks", // Зависит от nodes
		"nodes",           // Зависит от systems
		"products",        // Зависит от systems
		"systems",         // Зависит от projects
		"projects",        // Базовая таблица
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// Отключаем проверку внешних ключей
		if err := tx.Exec("SET session_replication_role = 'replica'").Error; err != nil {
			return fmt.Errorf("failed to disable FK checks: %w", err)
		}

		// Очищаем таблицы в правильном порядке
		for _, table := range tables {
			if tx.Migrator().HasTable(table) {
				if err := tx.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)).Error; err != nil {
					return fmt.Errorf("failed to truncate table %s: %w", table, err)
				}
				log.Printf("Table %s truncated", table)
			}
		}
		sequences := []string{
			"fb_variables_id_seq",
			"signals_id_seq",
			"function_blocks_id_seq",
			"nodes_id_seq",
			"products_id_seq",
			"systems_id_seq",
			"projects_id_seq",
		}
		for _, seq := range sequences {
			if err := tx.Exec(fmt.Sprintf("ALTER SEQUENCE %s RESTART WITH 1", seq)).Error; err != nil {
				log.Printf("Warning: failed to reset sequence %s: %v", seq, err)
				// Не прерываем выполнение, так как это не критично
			}
		}
		// Включаем проверку внешних ключей обратно
		if err := tx.Exec("SET session_replication_role = 'origin'").Error; err != nil {
			return fmt.Errorf("failed to enable FK checks: %w", err)
		}

		return nil
	})
}

func AddUniqueConstraints(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if !tx.Migrator().HasTable("products") || !tx.Migrator().HasTable("nodes") {
			return errors.New("required tables don't exist")
		}

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
		&models.System{},
	}

	for _, model := range modelsToMigrate {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to auto-migrate model %T: %w", model, err)
		}
	}
	return nil
}
