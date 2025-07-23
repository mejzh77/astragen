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
func (r *SystemRepository) GetAllSystemNames() ([]string, error) {
	var names []string
	result := r.db.Model(&models.System{}).Pluck("name", &names)
	if result.Error != nil {
		return nil, result.Error
	}
	return names, nil
}
func (r *SystemRepository) GetSystemByName(name string) (*models.System, error) {
	var system models.System
	if err := r.db.Where("name = ?", name).First(&system).Error; err != nil {
		return nil, fmt.Errorf("failed to get system: %w", err)
	}
	return &system, nil
}
func (r *SystemRepository) GetWithDetails(id string, system *models.System) error {
	return r.db.
		Preload("Nodes").
		Preload("Products").
		Preload("FunctionBlocks.Variables").
		First(system, id).Error
}
func (r *SystemRepository) GetSystemByTag(tag string) (*models.System, error) {
	var system models.System
	err := r.db.Where("tag = ?", tag).First(&system).Error
	return &system, err
}
