package sync

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mejzh77/astragen/configs/config"
	"github.com/mejzh77/astragen/pkg/models"
)

func (s *SyncService) syncNodesAndProductsFromSheets(ctx context.Context) error {
	nodes, err := s.loadNodesFromSheets(ctx)
	if err != nil {
		return fmt.Errorf("failed to load nodes: %w", err)
	}

	if err := s.syncNodes(nodes); err != nil {
		return fmt.Errorf("failed to save nodes: %w", err)
	}

	products, err := s.loadProductsFromSheets(ctx)
	if err != nil {
		return fmt.Errorf("failed to load products: %w", err)
	}

	if err := s.productRepo.BulkUpsert(products); err != nil {
		return fmt.Errorf("failed to save products: %w", err)
	}

	return nil
}

func (s *SyncService) GetNodeDetails(id string) (gin.H, error) {
	var node models.Node
	if err := s.nodeRepo.GetWithFBs(id, &node); err != nil {
		return nil, err
	}
	return node.ToDetailedAPI(), nil
}

func (s *SyncService) syncNodes(nodes []models.Node) error {
	for _, node := range nodes {
		if err := s.nodeRepo.SaveNodeWithSystems(&node); err != nil {
			return fmt.Errorf("failed to save node %s: %w", node.Name, err)
		}
	}
	return nil
}

func (s *SyncService) loadNodesFromSheets(ctx context.Context) ([]models.Node, error) {
	var sheetNodes []models.SheetNode

	if err := s.gsheets.Load(config.Cfg.SpreadsheetID, config.Cfg.NodeSheet, &sheetNodes); err != nil {
		return nil, fmt.Errorf("failed to load nodes: %w", err)
	}

	var nodes []models.Node
	for _, sn := range sheetNodes {
		node := models.Node{
			Name: sn.Name,
			Tag:  sn.Tag,
		}

		systemIDs := strings.Split(sn.Systems, ",")
		for _, sysID := range systemIDs {
			sysID = strings.TrimSpace(sysID)
			if sysID == "" {
				continue
			}

			system, err := s.systemRepo.GetSystemByName(sysID)
			if err != nil {
				return nil, fmt.Errorf("failed to get system %s: %w", sysID, err)
			}
			node.Systems = append(node.Systems, system)
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

func (s *SyncService) findBestNodeMatch(nodeName string, systemID uint) (*models.Node, error) {
	if node, err := s.nodeRepo.FindByName(nodeName); err == nil {
		return node, nil
	}
	if nodeName == "" {
		nodeName = "Общее"
	}
	nodes, err := s.nodeRepo.FindSimilarInSystem(nodeName)
	if err != nil {
		return nil, err
	}

	if len(nodes) > 0 {
		return &nodes[0], nil
	}

	newNode := &models.Node{
		Name:     nodeName,
		SystemID: &systemID,
	}
	if err := s.nodeRepo.Create(newNode); err != nil {
		return nil, err
	}

	return newNode, nil
}
