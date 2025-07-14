package repository

import (
	"fmt"

	"github.com/mejzh77/astragen/pkg/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type NodeRepository struct {
	db *gorm.DB
}

func NewNodeRepository(db *gorm.DB) *NodeRepository {
	return &NodeRepository{db: db}
}

// CreateOrUpdate создает или обновляет узел
func (r *NodeRepository) CreateOrUpdate(node *models.Node) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Пытаемся найти существующий узел
		var existing models.Node
		if err := tx.Where("name = ?", node.Name).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Создаем новый узел
				if err := tx.Create(node).Error; err != nil {
					return fmt.Errorf("failed to create node: %w", err)
				}
				return nil
			}
			return fmt.Errorf("failed to query node: %w", err)
		}

		// Обновляем существующий узел
		node.ID = existing.ID
		if err := tx.Save(node).Error; err != nil {
			return fmt.Errorf("failed to update node: %w", err)
		}

		return nil
	})
}

// GetByName возвращает узел по имени
func (r *NodeRepository) GetByName(name string) (*models.Node, error) {
	var node models.Node
	if err := r.db.Where("name = ?", name).First(&node).Error; err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}
	return &node, nil
}
func (r *NodeRepository) GetOrCreate(node *models.Node) error {
	return r.db.Where(models.Node{Name: node.Name}).FirstOrCreate(node).Error
}

// LinkToSystem связывает узел с системой
func (r *NodeRepository) LinkToSystem(node *models.Node, systemID uint) error {
	node.SystemID = &systemID
	return r.db.Model(node).Update("system_id", systemID).Error
}
func (r *NodeRepository) GetWithDetails(id string, node *models.Node) error {
	return r.db.
		Preload("System").
		First(node, id).Error
}

// В node_repository.go
func (r *NodeRepository) FindSimilarInSystem(name string, systemID uint) ([]models.Node, error) {
	var nodes []models.Node

	// Используем полнотекстовый поиск или similarity-функции
	// Пример для PostgreSQL:
	query := `SELECT * FROM nodes 
              WHERE system_id = ? 
              ORDER BY similarity(name, ?) DESC 
              LIMIT 3`

	if err := r.db.Raw(query, systemID, name).Scan(&nodes).Error; err != nil {
		return nil, err
	}

	return nodes, nil
}

// BulkUpsert создает или обновляет узлы пачкой
func (r *NodeRepository) BulkUpsert(nodes []models.Node) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}, {Name: "system_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"tag"}),
	}).Create(&nodes).Error
}

// FindByNameAndSystem ищет узел по точному совпадению имени и системы
func (r *NodeRepository) FindByNameAndSystem(name string, systemID uint) (*models.Node, error) {
	var node models.Node
	err := r.db.Where("name = ? AND system_id = ?", name, systemID).First(&node).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// Create создает новый узел
func (r *NodeRepository) Create(node *models.Node) error {
	return r.db.Create(node).Error
}
