package repository

import (
	"fmt"

	"github.com/mejzh77/astragen/pkg/models"
	"gorm.io/gorm"
)

type SystemRepository struct {
	db *gorm.DB
}

func NewSystemRepository(db *gorm.DB) *SystemRepository {
	return &SystemRepository{db: db}
}

func (r *SystemRepository) LinkSystemToProject(systemName string, projectID uint) (*models.System, error) {
	system := &models.System{
		Name:      systemName,
		ProjectID: projectID,
	}

	result := r.db.Where(models.System{
		Name:      systemName,
		ProjectID: projectID,
	}).FirstOrCreate(system)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to link system to project: %w", result.Error)
	}

	return system, nil
}

func (r *SystemRepository) GetSystemByName(name string) (*models.System, error) {
	var system models.System
	if err := r.db.Where("name = ?", name).First(&system).Error; err != nil {
		return nil, fmt.Errorf("failed to get system: %w", err)
	}
	return &system, nil
}
