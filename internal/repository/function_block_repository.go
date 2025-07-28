// function_block_repository.go
package repository

import (
	"fmt"
	"log"
	"strings"

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

// Добавляем новые методы в FunctionBlockRepository
func (r *FunctionBlockRepository) GetFiltered(system, cdsType, node string) ([]*models.FunctionBlock, error) {
	query := r.db.Preload("Variables").Preload("System").Preload("Node")

	if system != "" {
		query = query.
			Joins("JOIN systems ON systems.id = function_blocks.system_id").
			Where("systems.name = ?", system)
	}

	if cdsType != "" {
		query = query.Where("cds_type = ?", cdsType)
	}

	if node != "" {
		query = query.Joins("JOIN nodes ON nodes.id = function_blocks.node_id").
			Where("nodes.name = ?", node)
	}
	var fbs []*models.FunctionBlock
	if err := query.Find(&fbs).Error; err != nil {
		return nil, fmt.Errorf("failed to get filtered FBs: %w", err)
	}

	return fbs, nil
}

func (r *FunctionBlockRepository) GetAllCDSTypes() ([]string, error) {
	var types []string
	err := r.db.Model(&models.FunctionBlock{}).
		Distinct().
		Pluck("cds_type", &types).
		Error

	if err != nil {
		return nil, fmt.Errorf("failed to get CDS types: %w", err)
	}

	return types, nil
}

func (r *FunctionBlockRepository) GetBySystem(systemID uint) ([]*models.FunctionBlock, error) {
	var fbs []*models.FunctionBlock
	err := r.db.
		Preload("Variables").
		Where("system_id = ?", systemID).
		Find(&fbs).
		Error

	if err != nil {
		return nil, fmt.Errorf("failed to get FBs by system: %w", err)
	}

	return fbs, nil
}

func (r *FunctionBlockRepository) GetByNode(nodeID uint) ([]*models.FunctionBlock, error) {
	var fbs []*models.FunctionBlock
	err := r.db.
		Preload("Variables").
		Where("node_id = ?", nodeID).
		Find(&fbs).
		Error

	if err != nil {
		return nil, fmt.Errorf("failed to get FBs by node: %w", err)
	}

	return fbs, nil
}
func (r *FunctionBlockRepository) SyncInputsFromSignals(signals []models.Signal) error {
	fbConfigs := config.Cfg.FunctionBlocks

	return r.db.Transaction(func(tx *gorm.DB) error {
		// Первый проход: создаем/обновляем FB и переменные
		fbCache := make(map[string]*models.FunctionBlock)
		var fbTags []string

		for _, signal := range signals {
			fbConfig, exists := fbConfigs[signal.SignalType]
			if !exists {
				continue
			}
			var isInput bool
			for _, v := range fbConfig.In {
				if v == "address" {
					isInput = true
					break
				}
			}
			if !isInput {
				continue
			}

			fb, err := models.ParseFromSignal(signal, config.Cfg.AddressTemplate[signal.SignalType])
			if err != nil {
				continue
			}

			fb.CdsType = signal.SignalType

			if cachedFB, ok := fbCache[fb.Tag]; ok {
				fb = cachedFB
			} else {
				if err := tx.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "tag"}},
					DoUpdates: clause.AssignmentColumns([]string{"cds_type", "system_id", "node_id", "updated_at", "primary"}),
				}).Create(fb).Error; err != nil {
					return fmt.Errorf("failed to upsert FB %s: %w", fb.Tag, err)
				}
				fbCache[fb.Tag] = fb
				fbTags = append(fbTags, fb.Tag)
			}
		}

		// Второй проход: генерация ST-кода для всех FB с переменными
		// Второй проход: генерация контента для всех FB
		for _, fbTag := range fbTags {
			fb := fbCache[fbTag]
			fbConfig := fbConfigs[fb.CdsType]

			// Генерируем контент
			if err := r.GenerateFBContent(fb, fbConfig, &config.Cfg.DefaultOPCItem); err != nil {
				return err
			}

			// Сохраняем обновленные поля
			if err := tx.Model(fb).Updates(map[string]interface{}{
				"declaration": fb.Declaration,
				"call":        fb.Call,
				"omx":         fb.OMX,
				"opc":         fb.OPC,
			}).Error; err != nil {
				return fmt.Errorf("failed to update FB content %s: %w", fb.Tag, err)
			}
		}

		return nil
	})
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
			invertedIn := make(map[string]bool)
			for _, v := range fbConfig.In {
				parts := strings.Split(v, ".")
				invertedIn[parts[0]] = true
			}

			invertedOut := make(map[string]bool)
			for _, v := range fbConfig.Out {
				parts := strings.Split(v, ".")
				invertedOut[parts[0]] = true
			}

			// Теперь проверяем за O(1)
			var direction string
			if invertedIn[funcAttr] {
				direction = "input"
			} else if invertedOut[funcAttr] {
				direction = "output"
			} else {
				continue
			}

			fb, variable, err := models.ParseFBFromSignal(signal, direction, config.Cfg.AddressTemplate[signal.SignalType])
			if err != nil {
				fmt.Printf("failed to parse FB %s: %v", signal.Tag, err)
				continue
			}

			fb.CdsType = signal.FB

			if cachedFB, ok := fbCache[fb.Tag]; ok {
				fb = cachedFB
			} else {
				if err := tx.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "tag"}},
					DoUpdates: clause.AssignmentColumns([]string{"cds_type", "system_id", "updated_at"}),
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
		// Второй проход: генерация контента для всех FB
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

			// Генерируем контент
			if err := r.GenerateFBContent(fb, fbConfig, &config.Cfg.DefaultOPCItem); err != nil {
				return err
			}

			// Сохраняем обновленные поля
			if err := tx.Model(fb).Updates(map[string]interface{}{
				"declaration": fb.Declaration,
				"call":        fb.Call,
				"omx":         fb.OMX,
				"opc":         fb.OPC,
			}).Error; err != nil {
				return fmt.Errorf("failed to update FB content %s: %w", fb.Tag, err)
			}
		}

		return nil
	})
}
func (r *FunctionBlockRepository) GenerateFBContent(fb *models.FunctionBlock, fbConfig config.FBConfig, opcTemplate *config.OPCItemTemplate) error {
	stDecl, err := fb.GenerateSTDecl()
	if err != nil {
		return fmt.Errorf("failed to generate ST declaration for FB %s: %w", fb.Tag, err)
	}
	fb.Declaration = stDecl

	// Генерация ST-кода
	stCode, err := fb.GenerateSTCode(fbConfig.Template, fbConfig.In, fbConfig.Out)
	if err != nil {
		return fmt.Errorf("failed to generate ST code for FB %s: %w", fb.Tag, err)
	}
	fb.Call = stCode

	// Генерация OMX
	omxCode, err := fb.GenerateOMX(fbConfig.OMX.Template, fbConfig.OMX.Attributes)
	if err != nil {
		return fmt.Errorf("failed to generate OMX for FB %s: %w", fb.Tag, err)
	}
	fb.OMX = omxCode

	// Генерация OPC
	opcData := models.OPCTemplate{
		Binding:    opcTemplate.Binding,
		Namespace:  opcTemplate.Namespace,
		BasePath:   opcTemplate.BasePath,
		NodePrefix: opcTemplate.NodePrefix,
		PathSuffix: fbConfig.OPC.Items,
	}
	opcCode, err := fb.GenerateOPC(opcData)
	if err != nil {
		return fmt.Errorf("failed to generate OPC for FB %s: %w", fb.Tag, err)
	}
	fb.OPC = opcCode

	return nil
}

type SignalWithFB struct {
	Signal models.Signal
	FB     models.FunctionBlock `gorm:"embedded"`
}

func (r *FunctionBlockRepository) UpdateAddresses() error {
	var signals []SignalWithFB
	result := r.db.Table("signals").
		Select("signals.*, fb.*").
		Joins("JOIN fb_variables v ON v.signal_tag = signals.tag").
		Joins("JOIN function_blocks fb ON v.fb_id = fb.id").
		Where("fb.primary = ?", true).
		Find(&signals)
	if result.Error != nil {
		return fmt.Errorf("ошибка получения сигналов: %v", result.Error)
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, s := range signals {
			newAddress, err := models.UpdateAddress(s.Signal, config.Cfg.AddressTemplate[s.Signal.SignalType])
			if err != nil {
				// Можно добавить логирование и продолжить
				log.Printf("Ошибка обновления адреса для сигнала %s: %v", s.Signal.Tag, err)
				continue
			}

			// 3. Обновляем только primary = true блоки
			result := tx.Model(&models.FunctionBlock{}).
				Where("id = ? AND primary = ?", s.FB.ID, true).
				Update("address", newAddress)

			if result.Error != nil {
				return fmt.Errorf("ошибка обновления FB %d: %v", s.FB.ID, result.Error)
			}
		}
		return nil
	})
}

// RegenerateAllImportFiles перегенерирует ST, OMX и OPC для всех функциональных блоков
func (r *FunctionBlockRepository) RegenerateAllImportFiles() (map[string]map[string]string, error) {
	// Получаем все функциональные блоки с переменными
	if err := r.UpdateAddresses(); err != nil {
		return nil, fmt.Errorf("failed to get all FBs: %w", err)
	}
	var fbs []*models.FunctionBlock
	if err := r.db.Preload("Variables").Find(&fbs).Error; err != nil {
		return nil, fmt.Errorf("failed to get all FBs: %w", err)
	}

	result := make(map[string]map[string]string)
	fbConfigs := config.Cfg.FunctionBlocks

	for _, fb := range fbs {
		fbConfig, exists := fbConfigs[fb.CdsType]
		if !exists {
			continue
		}

		// Генерируем содержимое
		if err := r.GenerateFBContent(fb, fbConfig, &config.Cfg.DefaultOPCItem); err != nil {
			return nil, err
		}

		// Формируем результат
		result[fb.Tag] = map[string]string{
			"ST":  fb.Call,
			"OMX": fb.OMX,
			"OPC": fb.OPC,
		}

		// Обновляем в БД (опционально)
		if err := r.db.Model(fb).Updates(map[string]interface{}{
			"call": fb.Call,
			"omx":  fb.OMX,
			"opc":  fb.OPC,
		}).Error; err != nil {
			return nil, fmt.Errorf("failed to update FB %s: %w", fb.Tag, err)
		}
	}

	return result, nil
}
func (r *FunctionBlockRepository) GetWithDetails(id string, fb *models.FunctionBlock) error {
	return r.db.
		Preload("Variables").
		First(fb, id).Error
}
