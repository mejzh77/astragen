package sync

import (
	"context"
	"fmt"
	"reflect"

	"github.com/mejzh77/astragen/configs/config"
	"github.com/mejzh77/astragen/internal/gsheets"
	"github.com/mejzh77/astragen/internal/repository"
	"github.com/mejzh77/astragen/pkg/models"
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
	// 1. Синхронизация проектов и систем
	project, err := s.projectRepo.GetOrCreateDefaultProject()
	if err != nil {
		return fmt.Errorf("failed to sync projects: %w", err)
	}

	// Синхронизация систем из конфига
	for _, sysConfig := range config.Cfg.Systems {
		_, err := s.systemRepo.LinkSystemToProject(sysConfig, project.ID)
		if err != nil {
			return fmt.Errorf("failed to sync system %s: %w", sysConfig, err)
		}
	}

	// 2. Синхронизация сигналов
	allSignals, err := s.loadSignalsFromSheets(ctx)
	if err != nil {
		return fmt.Errorf("failed to load signals: %w", err)
	}

	if err := s.signalRepo.SaveSignals(allSignals, false); err != nil {
		return fmt.Errorf("failed to save signals: %w", err)
	}

	// 3. Синхронизация функциональных блоков
	if err := s.fbRepo.SyncFBFromSignals(allSignals); err != nil {
		return fmt.Errorf("failed to sync function blocks: %w", err)
	}

	// 4. Синхронизация узлов и продуктов
	if err := s.syncNodesAndProducts(allSignals); err != nil {
		return fmt.Errorf("failed to sync nodes and products: %w", err)
	}

	return nil
}

func (s *SyncService) syncNodesAndProducts(signals []models.Signal) error {
	// Собираем уникальные узлы и продукты из сигналов
	nodeSystemMap := make(map[string]string)    // nodeName -> systemName
	productSystemMap := make(map[string]string) // productName -> systemName

	for _, signal := range signals {
		if signal.NodeRef != "" {
			nodeSystemMap[signal.NodeRef] = signal.System
		}
		if signal.Product != nil && signal.Product.Name != "" {
			productSystemMap[signal.Product.Name] = signal.System
		}
	}

	// Создаем/обновляем узлы
	for nodeName, systemName := range nodeSystemMap {
		system, err := s.systemRepo.GetSystemByName(systemName)
		if err != nil {
			return fmt.Errorf("failed to get system for node %s: %w", nodeName, err)
		}

		// Получаем или создаем узел
		node := &models.Node{Name: nodeName}
		if err := s.nodeRepo.GetOrCreate(node); err != nil {
			return fmt.Errorf("failed to get or create node %s: %w", nodeName, err)
		}

		// Связываем с системой
		if err := s.nodeRepo.LinkToSystem(node, system.ID); err != nil {
			return fmt.Errorf("failed to link node %s to system: %w", nodeName, err)
		}
	}

	// Создаем/обновляем изделия
	for productName, systemName := range productSystemMap {
		system, err := s.systemRepo.GetSystemByName(systemName)
		if err != nil {
			return fmt.Errorf("failed to get system for product %s: %w", productName, err)
		}

		// Получаем или создаем продукт
		product := &models.Product{Name: productName}
		if err := s.productRepo.GetOrCreate(product); err != nil {
			return fmt.Errorf("failed to get or create product %s: %w", productName, err)
		}

		// Связываем с системой
		if err := s.productRepo.LinkToSystem(product, system.ID); err != nil {
			return fmt.Errorf("failed to link product %s to system: %w", productName, err)
		}
	}

	return nil
}

// SyncProducts синхронизирует продукты (изделия) из сигналов
// Каждый продукт принадлежит только одной системе
func (s *SyncService) SyncProducts(signals []models.Signal) error {
	// Собираем информацию о продуктах и их системах
	// Если продукт встречается в нескольких системах, берем первую попавшуюся
	productSystemMap := make(map[string]string) // productName -> systemName

	for _, signal := range signals {
		if signal.Product != nil && signal.Product.Name != "" {
			// Если продукт уже есть в мапе, не перезаписываем систему
			if _, exists := productSystemMap[signal.Product.Name]; !exists {
				productSystemMap[signal.Product.Name] = signal.System
			}
		}
	}

	// Создаем/обновляем продукты
	for productName, systemName := range productSystemMap {
		// Получаем систему
		system, err := s.systemRepo.GetSystemByName(systemName)
		if err != nil {
			return fmt.Errorf("failed to get system %s for product %s: %w",
				systemName, productName, err)
		}

		// Создаем или получаем продукт
		product := &models.Product{
			Name:     productName,
			SystemID: &system.ID, // Устанавливаем связь с системой
		}

		// Используем Upsert (создать или обновить)
		if err := s.productRepo.Upsert(product); err != nil {
			return fmt.Errorf("failed to upsert product %s: %w", productName, err)
		}
	}

	return nil
}

// SyncFunctionBlocks синхронизирует функциональные блоки из сигналов
// Это обертка вокруг SyncFBFromSignals с добавлением логирования
func (s *SyncService) SyncFunctionBlocks(signals []models.Signal) error {
	if err := s.fbRepo.SyncFBFromSignals(signals); err != nil {
		return fmt.Errorf("failed to sync function blocks: %w", err)
	}
	return nil
}
func (s *SyncService) loadSignalsFromSheets(ctx context.Context) ([]models.Signal, error) {
	var allSignals []models.Signal

	for _, sheetCfg := range config.Cfg.Sheets {
		readRange, err := gsheets.GetRange(sheetCfg.SheetName, sheetCfg.Model, true)
		if err != nil {
			return nil, fmt.Errorf("failed to GetRange for sheet %s: %w", sheetCfg.SheetName, err)
		}

		rows, err := s.gsheets.ReadSheet(config.Cfg.SpreadsheetID, readRange)
		if err != nil {
			return nil, fmt.Errorf("failed to read sheet %s: %w", sheetCfg.SheetName, err)
		}

		signals, err := s.parseSheetData(rows, sheetCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to parse sheet %s: %w", sheetCfg.SheetName, err)
		}

		allSignals = append(allSignals, signals...)
	}

	return allSignals, nil
}

// SyncSystemsFromSignals создает системы на основе данных из сигналов
func (s *SyncService) SyncSystemsFromSignals(signals []models.Signal) error {
	// 1. Получаем или создаем проект по умолчанию
	project, err := s.projectRepo.GetOrCreateDefaultProject()
	if err != nil {
		return fmt.Errorf("failed to get default project: %w", err)
	}

	// 2. Собираем уникальные системы из сигналов
	systemNames := make(map[string]struct{})
	for _, signal := range signals {
		if signal.System != "" {
			systemNames[signal.System] = struct{}{}
		}
	}

	// 3. Создаем системы
	for systemName := range systemNames {
		_, err := s.systemRepo.LinkSystemToProject(systemName, project.ID)
		if err != nil {
			return fmt.Errorf("failed to create system %s: %w", systemName, err)
		}
	}

	return nil
}
func (s *SyncService) parseSheetData(rows [][]interface{}, sheetCfg config.SheetConfig) ([]models.Signal, error) {
	modelSlice := reflect.New(reflect.SliceOf(reflect.TypeOf(sheetCfg.Model).Elem()))
	if err := gsheets.Unmarshal(rows, modelSlice.Interface()); err != nil {
		return nil, fmt.Errorf("failed to unmarshal sheet data: %w", err)
	}

	var signals []models.Signal
	for i := 0; i < modelSlice.Elem().Len(); i++ {
		item := modelSlice.Elem().Index(i).Addr().Interface()
		var signal models.Signal

		switch v := item.(type) {
		case *models.DI:
			signal.FromDI(*v)
		case *models.AI:
			signal.FromAI(*v)
		case *models.DO:
			signal.FromDO(*v)
		case *models.AO:
			signal.FromAO(*v)
		}

		signal.SignalType = sheetCfg.SheetName
		signals = append(signals, signal)
	}

	return signals, nil
}
func (s *SyncService) SyncProjectsAndSystems() error {
	project, err := s.projectRepo.GetOrCreateDefaultProject()
	if err != nil {
		return fmt.Errorf("failed to sync projects: %w", err)
	}
	for _, sysConfig := range config.Cfg.Systems {
		_, err := s.systemRepo.LinkSystemToProject(sysConfig, project.ID)
		if err != nil {
			return fmt.Errorf("failed to sync system %s: %w", sysConfig, err)
		}
	}
	return nil
}

func (s *SyncService) LoadAndSaveSignals(ctx context.Context) ([]models.Signal, error) {
	signals, err := s.loadSignalsFromSheets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load signals: %w", err)
	}

	if err := s.signalRepo.SaveSignals(signals, false); err != nil {
		return nil, fmt.Errorf("failed to save signals: %w", err)
	}
	return signals, nil
}

func (s *SyncService) SyncNodes(signals []models.Signal) error {
	nodeSystems := make(map[string][]string) // nodeName -> []systemNames

	for _, signal := range signals {
		if signal.NodeRef != "" {
			if _, exists := nodeSystems[signal.NodeRef]; !exists {
				nodeSystems[signal.NodeRef] = []string{}
			}
			// Добавляем систему, если ее еще нет в списке
			if !contains(nodeSystems[signal.NodeRef], signal.System) {
				nodeSystems[signal.NodeRef] = append(nodeSystems[signal.NodeRef], signal.System)
			}
		}
	}

	for nodeName, systemNames := range nodeSystems {
		node := &models.Node{Name: nodeName}
		if err := s.nodeRepo.GetOrCreate(node); err != nil {
			return fmt.Errorf("failed to get/create node %s: %w", nodeName, err)
		}

		for _, systemName := range systemNames {
			system, err := s.systemRepo.GetSystemByName(systemName)
			if err != nil {
				return fmt.Errorf("failed to get system %s: %w", systemName, err)
			}

			if err := s.nodeRepo.LinkToSystem(node, system.ID); err != nil {
				return fmt.Errorf("failed to link node %s to system %s: %w",
					nodeName, systemName, err)
			}
		}
	}
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
