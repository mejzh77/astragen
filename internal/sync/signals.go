package sync

import (
	"context"
	"fmt"
	"reflect"

	"github.com/mejzh77/astragen/configs/config"
	"github.com/mejzh77/astragen/internal/gsheets"
	"github.com/mejzh77/astragen/pkg/models"
)

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

func (s *SyncService) LinkSignalsWithFuzzyMatching(signals []models.Signal) error {
	for i, signal := range signals {
		if signal.NodeRef == "" {
			continue
		}

		system, err := s.systemRepo.GetSystemByName(signal.SystemRef)
		if err != nil {

			fmt.Printf("failed to get system %s: %s", signal.SystemRef, err)
			continue
		}

		node, err := s.findBestNodeMatch(signal.NodeRef, system.ID)
		if err != nil {
			return fmt.Errorf("failed to find node for %s: %w", signal.NodeRef, err)
		}

		signals[i].NodeID = &node.ID
	}

	return s.signalRepo.UpdateSignalNodes(signals)
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

		for _, signal := range signals {
			if err := s.processSignalSystems(&signal); err != nil {
				return nil, fmt.Errorf("failed to link system %s: %w", signal.Tag, err)
			}
			if signal.SystemID != nil && signal.ProductID != nil {
				allSignals = append(allSignals, signal)
			}

		}
	}

	return allSignals, nil
}

func (s *SyncService) processSignalSystems(signal *models.Signal) error {
	if signal.SystemRef == "" {
		return nil
	}

	system, err := s.systemRepo.GetSystemByName(signal.SystemRef)
	if err != nil {
		//fmt.Printf("failed to get system %s: %w", signal.SystemRef, err)
		return nil
	}
	signal.SystemID = &system.ID
	signal.System = system
	product, err := s.productRepo.GetByName(signal.ProductRef)
	if err != nil {
		//fmt.Printf("failed to get system %s: %w", signal.SystemRef, err)
		return nil
	}
	signal.ProductID = &product.ID
	signal.Product = product
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
		case *models.DQ:
			signal.FromDQ(*v)
		case *models.AQ:
			signal.FromAQ(*v)
		}

		signal.SignalType = sheetCfg.SheetName
		signals = append(signals, signal)
	}

	return signals, nil
}
