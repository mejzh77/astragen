package sync

import (
	"context"
	"fmt"
	"os"

	"github.com/mejzh77/astragen/configs/config"
	"github.com/mejzh77/astragen/internal/gsheets"
	"github.com/mejzh77/astragen/internal/repository"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

type SyncService struct {
	gsheets     *gsheets.Service
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
		gsheets:     gsheets,
		projectRepo: repository.NewProjectRepository(db),
		signalRepo:  repository.NewSignalRepository(db),
		fbRepo:      repository.NewFunctionBlockRepository(db),
		nodeRepo:    repository.NewNodeRepository(db),
		productRepo: repository.NewProductRepository(db),
		systemRepo:  repository.NewSystemRepository(db),
	}
}

func (s *SyncService) RunFullSync(ctx context.Context) error {
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

	return nil
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

	// Обновляем поля конфига
	if db, ok := updates["db"].(map[string]interface{}); ok {
		if host, ok := db["host"].(string); ok {
			cfg.DB.Host = host
		}
		if port, ok := db["port"].(string); ok {
			cfg.DB.Port = port
		}
		if user, ok := db["user"].(string); ok {
			cfg.DB.User = user
		}
		if password, ok := db["password"].(string); ok {
			cfg.DB.Password = password
		}
		if name, ok := db["name"].(string); ok {
			cfg.DB.Name = name
		}
		if sslMode, ok := db["ssl_mode"].(string); ok {
			cfg.DB.SSLMode = sslMode
		}
	}

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

	if systems, ok := updates["systems"].([]interface{}); ok {
		cfg.Systems = make([]string, 0, len(systems))
		for _, sys := range systems {
			if s, ok := sys.(string); ok {
				cfg.Systems = append(cfg.Systems, s)
			}
		}
	}

	if fbs, ok := updates["function_blocks"].(map[string]interface{}); ok {
		for name, fbData := range fbs {
			if fb, ok := fbData.(map[string]interface{}); ok {
				// Обновляем или создаем новый function block
				currentFB, exists := cfg.FunctionBlocks[name]
				if !exists {
					currentFB = config.FBConfig{
						In:  make(map[string]string),
						Out: make(map[string]string),
					}
				}

				if template, ok := fb["st_template"].(string); ok {
					currentFB.Template = template
				}

				if in, ok := fb["in"].(map[string]interface{}); ok {
					for k, v := range in {
						if val, ok := v.(string); ok {
							currentFB.In[k] = val
						}
					}
				}

				if out, ok := fb["out"].(map[string]interface{}); ok {
					for k, v := range out {
						if val, ok := v.(string); ok {
							currentFB.Out[k] = val
						}
					}
				}

				if omx, ok := fb["omx"].(map[string]interface{}); ok {
					if template, ok := omx["template"].(string); ok {
						currentFB.OMX.Template = template
					}
					if attrs, ok := omx["attributes"].(map[string]interface{}); ok {
						for k, v := range attrs {
							if val, ok := v.(string); ok {
								currentFB.OMX.Attributes[k] = val
							}
						}
					}
				}

				if opc, ok := fb["opc"].(map[string]interface{}); ok {
					if items, ok := opc["items"].([]interface{}); ok {
						currentFB.OPC.Items = make([]string, 0, len(items))
						for _, item := range items {
							if s, ok := item.(string); ok {
								currentFB.OPC.Items = append(currentFB.OPC.Items, s)
							}
						}
					}
				}

				cfg.FunctionBlocks[name] = currentFB
			}
		}
	}

	// В конце метода UpdateConfig добавьте:
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Получите путь к конфиг-файлу из вашего приложения
	configPath := "config.yml" // или путь из переменных окружения/флагов
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
