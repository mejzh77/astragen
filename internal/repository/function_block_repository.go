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
func (r *FunctionBlockRepository) GetAllWithNodes(fbs *[]models.FunctionBlock) error {
	return r.db.Find(fbs).Error
}
func (r *FunctionBlockRepository) DebugCheckFunctionBlocks() {
	var count int64
	r.db.Model(&models.FunctionBlock{}).Count(&count)
	log.Printf("Total function blocks in DB: %d", count)

	var fbs []models.FunctionBlock
	r.db.Limit(5).Find(&fbs)
	log.Printf("Sample FBs: %+v", fbs)
}

func (r *FunctionBlockRepository) SyncFBFromSignals(signals []models.Signal) error {
	fbConfigs := config.Cfg.FunctionBlocks

	return r.db.Transaction(func(tx *gorm.DB) error {
		// Первый проход: создаем/обновляем FB и переменные
		fbCache := make(map[string]*models.FunctionBlock)
		var fbTags []string

		for _, signal := range signals {
			_, funcAttr, ok := models.ParseFBInfo(signal.Tag)
			if !ok || signal.FB == "" {
				continue
			}

			fbConfig, exists := fbConfigs[signal.FB]
			if !exists {
				continue
			}

			var direction string
			if _, isInput := fbConfig.In[funcAttr]; isInput {
				direction = "input"
			} else if _, isOutput := fbConfig.Out[funcAttr]; isOutput {
				direction = "output"
			} else {
				continue
			}

			fb, variable := models.ParseFBFromSignal(signal, direction)
			if fb == nil {
				continue
			}

			fb.CdsType = signal.FB

			if cachedFB, ok := fbCache[fb.Tag]; ok {
				fb = cachedFB
			} else {
				if err := tx.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "tag"}},
					DoUpdates: clause.AssignmentColumns([]string{"cds_type", "system_id", "node_id", "updated_at"}),
				}).Create(fb).Error; err != nil {
					return fmt.Errorf("failed to upsert FB %s: %w", fb.Tag, err)
				}
				fbCache[fb.Tag] = fb
				fbTags = append(fbTags, fb.Tag)
			}

			variable.FBID = fb.ID
			result := tx.Where(models.FBVariable{
				SignalTag: variable.SignalTag,
				FBID:      variable.FBID,
			}).FirstOrCreate(variable)

			if result.Error != nil {
				return fmt.Errorf("failed to upsert variable for FB %s: %w", fb.Tag, result.Error)
			}
		}

		// Второй проход: генерация ST-кода для всех FB с переменными
		for _, fbTag := range fbTags {
			fb := fbCache[fbTag]
			fbConfig := fbConfigs[fb.CdsType]

			// Загружаем все переменные для этого FB
			var variables []models.FBVariable
			if err := tx.Where("fb_id = ?", fb.ID).Find(&variables).Error; err != nil {
				return fmt.Errorf("failed to load variables for FB %s: %w", fb.Tag, err)
			}

			// Обновляем FB переменными
			fb.Variables = variables

			// Генерируем ST-код
			stCode, err := fb.GenerateSTCode(
				fbConfig.Template,
				fbConfig.In,
				fbConfig.Out,
			)
			if err != nil {
				return fmt.Errorf("failed to generate ST code for FB %s: %w", fb.Tag, err)
			}
			// Обновляем поле Call
			if err := tx.Model(fb).Update("call", stCode).Error; err != nil {
				return fmt.Errorf("failed to update FB call %s: %w", fb.Tag, err)
			}
			omxCode, err := fb.GenerateOMX(fbConfig.OMX.Template, fbConfig.OMX.Attributes)
			if err != nil {
				return fmt.Errorf("failed to generate OMX for FB %s: %w", fb.Tag, err)
			}
			// Обновляем поле Call
			if err := tx.Model(fb).Update("omx", omxCode).Error; err != nil {
				return fmt.Errorf("failed to update FB call %s: %w", fb.Tag, err)
			}
			opcData := models.OPCTemplate{
				Binding:    config.Cfg.DefaultOPCItem.Binding,
				Namespace:  config.Cfg.DefaultOPCItem.Namespace,
				BasePath:   config.Cfg.DefaultOPCItem.BasePath,
				NodePrefix: config.Cfg.DefaultOPCItem.NodePrefix,
				PathSuffix: fbConfig.OPC.Items,
			}
			opcCode, err := fb.GenerateOPC(opcData)
			if err != nil {
				return fmt.Errorf("failed to generate OMX for FB %s: %w", fb.Tag, err)
			}
			// Обновляем поле Call
			if err := tx.Model(fb).Update("opc", opcCode).Error; err != nil {
				return fmt.Errorf("failed to update FB call %s: %w", fb.Tag, err)
			}
		}

		return nil
	})
}

func (r *FunctionBlockRepository) GetWithDetails(id string, fb *models.FunctionBlock) error {
	return r.db.
		Preload("Variables").
		First(fb, id).Error
}
