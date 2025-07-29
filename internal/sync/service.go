package sync

import (
	"context"
	"fmt"
	"github.com/mejzh77/astragen/pkg/models"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/mejzh77/astragen/configs/config"
	"github.com/mejzh77/astragen/internal/gsheets"
	"github.com/mejzh77/astragen/internal/repository"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

type SyncService struct {
	gsRead      *gsheets.Service
	gsWrite     *gsheets.WriteService
	projectRepo *repository.ProjectRepository
	signalRepo  *repository.SignalRepository
	fbRepo      *repository.FunctionBlockRepository
	nodeRepo    *repository.NodeRepository
	productRepo *repository.ProductRepository
	systemRepo  *repository.SystemRepository
}

func NewSyncService(
	gsheets *gsheets.Service,
	db *gorm.DB,
) *SyncService {
	return &SyncService{
		gsRead:      gsheets,
		projectRepo: repository.NewProjectRepository(db),
		signalRepo:  repository.NewSignalRepository(db),
		fbRepo:      repository.NewFunctionBlockRepository(db),
		nodeRepo:    repository.NewNodeRepository(db),
		productRepo: repository.NewProductRepository(db),
		systemRepo:  repository.NewSystemRepository(db),
	}
}

func (s *SyncService) SetWriteService(sheetsService *gsheets.WriteService) {
	s.gsWrite = sheetsService
}

func (s *SyncService) SetReadService(sheetsService *gsheets.Service) {
	s.gsRead = sheetsService
}

func (s *SyncService) RunFullSync(ctx context.Context) error {
	log.Println("Initializing services...")
	creds, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Failed to read credentials file: %v", err)
	}

	sheetsService, err := gsheets.NewService(ctx, creds)
	if err != nil {
		log.Fatalf("Failed to create Google Sheets service: %v", err)
	}
	// 4. Полная синхронизация с логированием
	log.Println("Starting full sync process...")

	s.SetReadService(sheetsService)
	writeService, err := gsheets.NewWriteService(ctx, creds)
	if err == nil {
		s.SetWriteService(writeService)
	}
	// 4.2. Синхронизация сигналов
	if err := s.SyncProjectsAndSystems(); err != nil {
		return fmt.Errorf("failed to sync projects and systems: %w", err)
	}

	if err := s.syncNodesAndProductsFromSheets(ctx); err != nil {
		return fmt.Errorf("failed to sync nodes and products: %w", err)
	}

	signals, err := s.LoadAndSaveSignals(ctx)
	if err != nil {
		return fmt.Errorf("failed to sync signals: %w", err)
	}

	if err := s.LinkSignalsWithFuzzyMatching(signals); err != nil {
		return fmt.Errorf("failed to link signals: %w", err)
	}

	if err := s.SyncFunctionBlocks(signals); err != nil {
		return fmt.Errorf("failed to sync function blocks: %w", err)
	}

	if err := s.LinkFunctionBlocksToNodes(); err != nil {
		return fmt.Errorf("failed to link function blocks: %w", err)
	}
	if err := s.SyncFunctionBlocksWithSheet(config.Cfg.SpreadsheetID, "FB"); err != nil {
		return fmt.Errorf("failed to sync function blocks: %w", err)
	}
	return nil
}

// SyncFunctionBlocksWithSheet синхронизирует функциональные блоки между Google Sheets и базой данных
func (s *SyncService) SyncFunctionBlocksWithSheet(spreadsheetID, sheetName string) error {
	// 1. Получаем функциональные блоки из Google Sheets
	var sheetFBs []models.SheetFB
	err := s.gsRead.Load(spreadsheetID, sheetName, &sheetFBs)
	if err != nil {
		return fmt.Errorf("failed to get function blocks from sheet: %w", err)
	}
	var dbFBs []models.FunctionBlock
	// 2. Получаем функциональные блоки из базы данных
	err = s.fbRepo.GetAll(&dbFBs)
	if err != nil {
		return fmt.Errorf("failed to get function blocks from db: %w", err)
	}

	// 3. Создаем карты для быстрого поиска
	sheetFBMap := make(map[string]models.SheetFB)
	for _, fb := range sheetFBs {
		sheetFBMap[fb.Tag] = fb
	}

	dbFBMap := make(map[string]models.FunctionBlock)
	for _, fb := range dbFBs {
		dbFBMap[fb.Tag] = fb
	}

	// 4. Синхронизация: Sheet -> DB
	var sheetToDBUpdates []models.SheetFB
	for tag, sheetFB := range sheetFBMap {
		if dbFB, exists := dbFBMap[tag]; exists {
			// Обновляем существующую запись в БД, если есть изменения
			if s.fbNeedsUpdate(dbFB, sheetFB) {
				sheetToDBUpdates = append(sheetToDBUpdates, sheetFB)
			}
		} else {
			// Добавляем новую запись в БД
			sheetToDBUpdates = append(sheetToDBUpdates, sheetFB)
		}
	}

	if len(sheetToDBUpdates) > 0 {
		if err := s.updateDBFunctionBlocks(sheetToDBUpdates); err != nil {
			return fmt.Errorf("failed to update db function blocks: %w", err)
		}
		log.Printf("Updated %d function blocks in DB from sheet", len(sheetToDBUpdates))
	}

	// 5. Синхронизация: DB -> Sheet (полная перезапись листа)
	if len(dbFBMap) > 0 {
		// Преобразуем все функциональные блоки в формат SheetFB
		var allSheetFBs []models.SheetFB
		for _, fb := range dbFBMap {
			if fb.Primary {
				continue
			}
			var sys, descr, name string
			if fb.System != nil {
				sys = fb.System.Name
			} else {
				sys = "--"
			}
			if fb.Name != sheetFBMap[fb.Tag].Name {
				name = sheetFBMap[fb.Tag].Name
			}
			if fb.Description != sheetFBMap[fb.Tag].Description {
				descr = sheetFBMap[fb.Tag].Description
			}
			allSheetFBs = append(allSheetFBs, models.SheetFB{
				Name:        name,
				Tag:         fb.Tag,
				Description: descr,
				CdsType:     fb.CdsType,
				System:      sys,
			})
		}

		// Используем метод Save для полной перезаписи листа
		if err := s.gsWrite.Save(spreadsheetID, sheetName, allSheetFBs); err != nil {
			return fmt.Errorf("failed to save function blocks to sheet: %w", err)
		}
		log.Printf("Saved %d function blocks to sheet", len(allSheetFBs))
	}

	return nil
}

// updateDBFunctionBlocks обновляет функциональные блоки в БД
func (s *SyncService) updateDBFunctionBlocks(sheetFBs []models.SheetFB) error {
	for _, sheetFB := range sheetFBs {
		if sheetFB.Name == "" {
			continue
		}
		fb := models.FunctionBlock{
			Tag:         sheetFB.Tag,
			Name:        sheetFB.Name,
			Description: sheetFB.Description,
		}

		if err := s.fbRepo.Upsert(fb); err != nil {
			return fmt.Errorf("failed to upsert function block %s: %w", sheetFB.Tag, err)
		}
	}
	return nil
}

// fbNeedsUpdate проверяет, нужно ли обновлять запись в БД
func (s *SyncService) fbNeedsUpdate(dbFB models.FunctionBlock, sheetFB models.SheetFB) bool {
	return dbFB.Name != sheetFB.Name || dbFB.Description != sheetFB.Description
}

// Добавьте эти методы в SyncService
func (s *SyncService) GetConfig() (map[string]interface{}, error) {
	// Преобразуем структуру конфига в map для удобства работы с фронтендом
	cfg := config.Cfg
	if cfg == nil {
		return nil, fmt.Errorf("config not loaded")
	}

	// Конвертируем структуру в map
	configMap := make(map[string]interface{})

	// Database
	configMap["db"] = map[string]interface{}{
		"host":     cfg.DB.Host,
		"port":     cfg.DB.Port,
		"user":     cfg.DB.User,
		"password": cfg.DB.Password,
		"name":     cfg.DB.Name,
		"ssl_mode": cfg.DB.SSLMode,
	}

	// Spreadsheet
	configMap["spreadsheet_id"] = cfg.SpreadsheetID
	configMap["nodesheet"] = cfg.NodeSheet
	configMap["productsheet"] = cfg.ProductSheet
	configMap["update"] = cfg.Update

	// Systems
	configMap["systems"] = cfg.Systems
	addresses := make(map[string]interface{})
	for t, s := range cfg.AddressTemplate {
		addresses[t] = s
	}
	configMap["signal_addresses"] = addresses
	// Function blocks
	fbs := make(map[string]interface{})
	for name, fb := range cfg.FunctionBlocks {
		fbs[name] = map[string]interface{}{
			"st_template": fb.Template,
			"in":          fb.In,
			"out":         fb.Out,
			"omx": map[string]interface{}{
				"template":   fb.OMX.Template,
				"attributes": fb.OMX.Attributes,
			},
			"opc": map[string]interface{}{
				"items": fb.OPC.Items,
			},
		}
	}
	configMap["function_blocks"] = fbs

	return configMap, nil
}

func (s *SyncService) UpdateConfig(updates map[string]interface{}) error {
	cfg := config.Cfg
	if cfg == nil {
		return fmt.Errorf("config not loaded")
	}

	// Обновляем простые поля
	if spreadsheetID, ok := updates["spreadsheet_id"].(string); ok {
		cfg.SpreadsheetID = spreadsheetID
	}
	if nodesheet, ok := updates["nodesheet"].(string); ok {
		cfg.NodeSheet = nodesheet
	}
	if productsheet, ok := updates["productsheet"].(string); ok {
		cfg.ProductSheet = productsheet
	}
	if update, ok := updates["update"].(bool); ok {
		cfg.Update = update
	}

	// Обработка систем (список)
	if systems, ok := updates["systems"].([]interface{}); ok {
		cfg.Systems = make([]string, 0, len(systems))
		for _, sys := range systems {
			if s, ok := sys.(string); ok {
				cfg.Systems = append(cfg.Systems, s)
			}
		}
	}
	if addresses, ok := updates["signal_addresses"].(map[string]interface{}); ok {
		cfg.AddressTemplate = make(map[string]string)
		for key, addr := range addresses {
			if s, ok := addr.(string); ok {
				cfg.AddressTemplate[key] = s
			}
		}
	}
	// Инициализация DB, если она nil
	if cfg.DB == nil {
		cfg.DB = &config.DatabaseConfig{
			Host:     "",
			Port:     "",
			User:     "",
			Password: "",
			Name:     "",
			SSLMode:  "disable",
		}
	}

	// Обработка функциональных блоков
	if fbs, ok := updates["function_blocks"].(map[string]interface{}); ok {
		if cfg.FunctionBlocks == nil {
			cfg.FunctionBlocks = make(map[string]config.FBConfig)
		}

		for name, fbData := range fbs {
			if fb, ok := fbData.(map[string]interface{}); ok {
				currentFB := cfg.FunctionBlocks[name]

				// Обновляем шаблон
				if template, ok := fb["st_template"].(string); ok {
					currentFB.Template = template
				}

				// Обработка входов (in)
				if in, ok := fb["in"].(map[string]interface{}); ok {
					if currentFB.In == nil {
						currentFB.In = make(map[string]string)
					}
					syncMap(currentFB.In, in)
				} else {
					// Если в обновлении нет in, очищаем существующие значения
					currentFB.In = make(map[string]string)
				}

				// Обработка выходов (out)
				if out, ok := fb["out"].(map[string]interface{}); ok {
					if currentFB.Out == nil {
						currentFB.Out = make(map[string]string)
					}
					syncMap(currentFB.Out, out)
				} else {
					currentFB.Out = make(map[string]string)
				}

				// Обработка OMX
				if omx, ok := fb["omx"].(map[string]interface{}); ok {
					if template, ok := omx["template"].(string); ok {
						currentFB.OMX.Template = template
					}
					if attrs, ok := omx["attributes"].(map[string]interface{}); ok {
						if currentFB.OMX.Attributes == nil {
							currentFB.OMX.Attributes = make(map[string]string)
						}
						syncMap(currentFB.OMX.Attributes, attrs)
					} else {
						currentFB.OMX.Attributes = make(map[string]string)
					}
				}

				// Обработка OPC
				if opc, ok := fb["opc"].(map[string]interface{}); ok {
					if items, ok := opc["items"].([]interface{}); ok {
						currentFB.OPC.Items = make([]string, 0, len(items))
						for _, item := range items {
							if s, ok := item.(string); ok {
								currentFB.OPC.Items = append(currentFB.OPC.Items, s)
							}
						}
					} else {
						currentFB.OPC.Items = nil
					}
				}

				cfg.FunctionBlocks[name] = currentFB
			}
		}
	}

	// Сохраняем конфиг
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := "config.yml" // или путь из переменных окружения/флагов
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Вспомогательная функция для синхронизации карт
func syncMap(dest map[string]string, src map[string]interface{}) {
	// Сначала удаляем ключи, которых нет в src
	for key := range dest {
		if _, exists := src[key]; !exists {
			delete(dest, key)
		}
	}

	// Затем добавляем/обновляем значения
	for key, val := range src {
		if strVal, ok := val.(string); ok {
			dest[key] = strVal
		}
	}
}

// Вспомогательная функция для обновления структур через reflection
func updateMapValues(src map[string]interface{}, dest interface{}) {
	val := reflect.ValueOf(dest).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Получаем тег yaml для имени поля
		yamlTag := fieldType.Tag.Get("yaml")
		if yamlTag == "" {
			yamlTag = strings.ToLower(fieldType.Name)
		}

		if srcVal, ok := src[yamlTag]; ok {
			switch field.Kind() {
			case reflect.String:
				if strVal, ok := srcVal.(string); ok {
					field.SetString(strVal)
				}
			case reflect.Bool:
				if boolVal, ok := srcVal.(bool); ok {
					field.SetBool(boolVal)
				}
			case reflect.Map:
				if srcMap, ok := srcVal.(map[string]interface{}); ok {
					destMap := field.Interface().(map[string]string)
					syncMap(destMap, srcMap)
				}
			}
		}
	}
}
