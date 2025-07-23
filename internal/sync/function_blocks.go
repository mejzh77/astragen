package sync

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/mejzh77/astragen/pkg/models"
)

func (s *SyncService) GetFunctionBlockDetails(id string) (gin.H, error) {
	var fb models.FunctionBlock
	if err := s.fbRepo.GetWithDetails(id, &fb); err != nil {
		return nil, fmt.Errorf("failed to get function block details: %w", err)
	}
	return fb.ToDetailedAPI(), nil
}

func (s *SyncService) LinkFunctionBlocksToNodes() error {
	var fbs []models.FunctionBlock
	if err := s.fbRepo.GetAllWithNodes(&fbs); err != nil {
		return fmt.Errorf("failed to get function blocks: %w", err)
	}
	for _, fb := range fbs {

		var node models.Node
		if fb.NodeID == nil {
			continue
		}
		err := s.nodeRepo.GetByID(*fb.NodeID, &node)
		if err != nil {
			log.Printf("Warning: failed to find node for FB %s: %v", fb.Tag, err)
			continue
		}

		if err := s.nodeRepo.LinkFunctionBlock(&node, &fb); err != nil {
			log.Printf("Warning: failed to link FB %s to node %s: %v", fb.Tag, node.Name, err)
			continue
		}
	}

	return nil
}

func (s *SyncService) SyncFunctionBlocks(signals []models.Signal) error {
	if err := s.fbRepo.SyncInputsFromSignals(signals); err != nil {
		return fmt.Errorf("failed to sync function blocks: %w", err)
	}
	if err := s.fbRepo.SyncFBFromSignals(signals); err != nil {
		return fmt.Errorf("failed to sync function blocks: %w", err)
	}
	return nil
}
func (s *SyncService) GetFilteredFunctionBlocks(system, cdsType, node string) ([]*models.FunctionBlock, error) {
	return s.fbRepo.GetFiltered(system, cdsType, node)
}

func (s *SyncService) GetAllCDSTypes() ([]string, error) {
	return s.fbRepo.GetAllCDSTypes()
}
func (s *SyncService) RegenerateAllImportFiles() (map[string]map[string]string, error) {
	return s.fbRepo.RegenerateAllImportFiles()
}
