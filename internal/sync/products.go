package sync

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/mejzh77/astragen/configs/config"
	"github.com/mejzh77/astragen/pkg/models"
)

func (s *SyncService) GetProductDetails(id string) (gin.H, error) {
	var product models.Product
	if err := s.productRepo.GetWithDetails(id, &product); err != nil {
		return nil, fmt.Errorf("failed to get product details: %w", err)
	}
	return product.ToDetailedAPI(), nil
}

func (s *SyncService) loadProductsFromSheets(ctx context.Context) ([]models.Product, error) {
	var sheetProducts []models.SheetProduct

	if err := s.gsRead.Load(config.Cfg.SpreadsheetID, config.Cfg.ProductSheet, &sheetProducts); err != nil {
		return nil, fmt.Errorf("failed to load products: %w", err)
	}

	var products []models.Product
	for _, sp := range sheetProducts {
		system, err := s.systemRepo.GetSystemByName(sp.System)
		if err != nil {
			continue
		}

		products = append(products, models.Product{
			PN:       sp.PN,
			Tag:      sp.Tag,
			Name:     sp.Name,
			Location: sp.Location,
			SystemID: &system.ID,
		})
	}

	return products, nil
}
