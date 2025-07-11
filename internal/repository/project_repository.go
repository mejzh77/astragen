// project_repository.go
package repository

import (
	"fmt"

	"github.com/mejzh77/astragen/pkg/models"
	"gorm.io/gorm"
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) GetOrCreateDefaultProject() (*models.Project, error) {
	defaultProject := &models.Project{
		Name:        "0101 Красный Бор",
		Description: "Автоматически созданный проект по умолчанию",
	}

	result := r.db.Where(models.Project{Name: defaultProject.Name}).FirstOrCreate(defaultProject)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get or create default project: %w", result.Error)
	}

	return defaultProject, nil
}

func (r *ProjectRepository) LinkSystemToProject(systemName string, systemType string) (*models.System, error) {
	project, err := r.GetOrCreateDefaultProject()
	if err != nil {
		return nil, err
	}

	system := &models.System{
		Name:      systemName,
		ProjectID: project.ID,
	}

	result := r.db.Where(models.System{
		Name:      systemName,
		ProjectID: project.ID,
	}).FirstOrCreate(system)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to link system to project: %w", result.Error)
	}

	return system, nil
}

func (r *ProjectRepository) LinkProductToSystem(productName string, systemName string) (*models.Product, error) {
	system, err := r.LinkSystemToProject(systemName, "product")
	if err != nil {
		return nil, err
	}

	product := &models.Product{
		Name:     productName,
		SystemID: &system.ID,
	}

	result := r.db.Where(models.Product{
		Name:     productName,
		SystemID: &system.ID,
	}).FirstOrCreate(product)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to link product to system: %w", result.Error)
	}

	return product, nil
}

func (r *ProjectRepository) LinkNodeToSystem(nodeName string, systemName string) (*models.Node, error) {
	system, err := r.LinkSystemToProject(systemName, "node")
	if err != nil {
		return nil, err
	}

	node := &models.Node{
		Name:     nodeName,
		SystemID: &system.ID,
	}

	result := r.db.Where(models.Node{
		Name:     nodeName,
		SystemID: &system.ID,
	}).FirstOrCreate(node)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to link node to system: %w", result.Error)
	}

	return node, nil
}
