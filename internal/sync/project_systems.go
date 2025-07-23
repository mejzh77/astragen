package sync

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/mejzh77/astragen/configs/config"
	"github.com/mejzh77/astragen/pkg/models"
)

func (s *SyncService) SyncProjectsAndSystems() error {
	project, err := s.projectRepo.GetOrCreateDefaultProject()
	if err != nil {
		return fmt.Errorf("failed to sync projects: %w", err)
	}

	for _, sysConfig := range config.Cfg.Systems {
		if _, err := s.systemRepo.LinkSystemToProject(sysConfig, project.ID); err != nil {
			return fmt.Errorf("failed to sync system %s: %w", sysConfig, err)
		}
	}
	return nil
}

func (s *SyncService) GetProjectDetails(id string) (gin.H, error) {
	var project models.Project
	if err := s.projectRepo.GetWithDetails(id, &project); err != nil {
		return nil, fmt.Errorf("failed to get project details: %w", err)
	}
	return project.ToDetailedAPI(), nil
}

func (s *SyncService) GetSystemDetails(id string) (gin.H, error) {
	var system models.System
	if err := s.systemRepo.GetWithDetails(id, &system); err != nil {
		return nil, fmt.Errorf("failed to get system details: %w", err)
	}
	return system.ToDetailedAPI(), nil
}

func (s *SyncService) GetProjectsWithHierarchy() ([]models.Project, error) {
	var projects []models.Project
	err := s.projectRepo.GetAllWithHierarchy(&projects)
	return projects, err
}

func (s *SyncService) GetTreeData() ([]gin.H, error) {
	var projects []models.Project
	if err := s.projectRepo.GetAllWithHierarchy(&projects); err != nil {
		return nil, fmt.Errorf("failed to load projects: %w", err)
	}

	var treeData []gin.H
	for _, p := range projects {
		treeData = append(treeData, p.ToAPI())
	}

	return treeData, nil
}
func (s *SyncService) GetAllSystems() ([]string, error) {
	return s.systemRepo.GetAllSystemNames()
}
