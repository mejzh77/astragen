package sync

import (
	"context"
	"fmt"

	"github.com/mejzh77/astragen/internal/gsheets"
	"github.com/mejzh77/astragen/internal/repository"
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

	if err := s.SyncFunctionBlocks(signals); err != nil {
		return fmt.Errorf("failed to sync function blocks: %w", err)
	}

	if err := s.LinkSignalsWithFuzzyMatching(signals); err != nil {
		return fmt.Errorf("failed to link signals: %w", err)
	}

	if err := s.LinkFunctionBlocksToNodes(); err != nil {
		return fmt.Errorf("failed to link function blocks: %w", err)
	}

	return nil
}
