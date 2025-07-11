// function_block_repository.go
package repository

import (
	"fmt"
	"log"

	"github.com/mejzh77/astragen/configs/config"
	"github.com/mejzh77/astragen/pkg/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FunctionBlockRepository struct {
	db *gorm.DB
}

func NewFunctionBlockRepository(db *gorm.DB) *FunctionBlockRepository {
	if err := createFunctionBlocksTables(db); err != nil {
		log.Fatalf("Failed to create function blocks tables: %v", err)
	}
	return &FunctionBlockRepository{db: db}
}

func createFunctionBlocksTables(db *gorm.DB) error {
	// Отключаем проверку внешних ключей для безопасного создания
	if err := db.Exec("SET CONSTRAINTS ALL DEFERRED").Error; err != nil {
		return err
	}

	tables := []string{
		`CREATE TABLE IF NOT EXISTS function_blocks (
			id SERIAL PRIMARY KEY,
			tag VARCHAR(255) NOT NULL UNIQUE,
			system VARCHAR(100),
			cds_type VARCHAR(50),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP WITH TIME ZONE
		)`,

		`CREATE TABLE IF NOT EXISTS fb_variables (
			id SERIAL PRIMARY KEY,
			fb_id INTEGER NOT NULL,
			direction VARCHAR(10) NOT NULL CHECK (direction IN ('input', 'output')),
			signal_tag VARCHAR(255) NOT NULL,
			func_attr VARCHAR(100) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP WITH TIME ZONE,
			CONSTRAINT fk_fb FOREIGN KEY(fb_id) REFERENCES function_blocks(id) ON DELETE CASCADE,
			CONSTRAINT fk_signal FOREIGN KEY(signal_tag) REFERENCES signals(tag) ON DELETE SET NULL
		)`,

		`CREATE INDEX IF NOT EXISTS idx_fb_variables_fb_id ON fb_variables(fb_id)`,
		`CREATE INDEX IF NOT EXISTS idx_fb_variables_signal_tag ON fb_variables(signal_tag)`,
	}

	for _, table := range tables {
		if err := db.Exec(table).Error; err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

func (r *FunctionBlockRepository) GetFBWithVariables(tag string) (*models.FunctionBlock, error) {
	var fb models.FunctionBlock
	err := r.db.Preload("Variables", func(db *gorm.DB) *gorm.DB {
		return db.Order("fb_variables.direction DESC, fb_variables.name") // Сначала inputs, потом outputs
	}).Where("tag = ?", tag).First(&fb).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get FB with variables: %w", err)
	}
	return &fb, nil
}

func (r *FunctionBlockRepository) SyncFBFromSignals(signals []models.Signal) error {
	// Собираем информацию о направлениях переменных из конфига
	varDirections := make(map[string]string) // signalTag -> direction
	for _, fbConfig := range config.Cfg.FunctionBlocks {
		for _, inVar := range fbConfig.In {
			varDirections[inVar] = "input"
		}
		for _, outVar := range fbConfig.Out {
			varDirections[outVar] = "output"
		}
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		fbCache := make(map[string]*models.FunctionBlock)

		for _, signal := range signals {
			_, funcAttr, _ := models.ParseFBInfo(signal.Tag)
			direction, exists := varDirections[funcAttr]
			if !exists {
				continue // Пропускаем сигналы не из конфига
			}

			fb, variable := models.ParseFBFromSignal(signal, direction)
			if fb == nil {
				continue
			}

			// Используем кэш FB, чтобы не дублировать
			if cachedFB, ok := fbCache[fb.Tag]; ok {
				fb = cachedFB
			} else {
				// Создаем/обновляем FB
				if err := tx.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "tag"}},
					DoUpdates: clause.AssignmentColumns([]string{"system", "cds_type", "updated_at"}),
				}).Create(fb).Error; err != nil {
					return fmt.Errorf("failed to upsert FB %s: %w", fb.Tag, err)
				}
				fbCache[fb.Tag] = fb
			}

			// Добавляем переменную
			variable.FBID = fb.ID
			if err := tx.Create(variable).Error; err != nil {
				return fmt.Errorf("failed to create variable for FB %s: %w", fb.Tag, err)
			}
		}
		return nil
	})
}
