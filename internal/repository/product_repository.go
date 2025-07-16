package repository

import (
	"errors"
	"fmt"

	"github.com/mejzh77/astragen/pkg/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// CreateOrUpdate создает или обновляет изделие
func (r *ProductRepository) CreateOrUpdate(product *models.Product) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Пытаемся найти существующее изделие
		var existing models.Product
		if err := tx.Where("name = ?", product.Name).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Создаем новое изделие
				if err := tx.Create(product).Error; err != nil {
					return fmt.Errorf("failed to create product: %w", err)
				}
				return nil
			}
			return fmt.Errorf("failed to query product: %w", err)
		}

		// Обновляем существующее изделие
		product.ID = existing.ID
		if err := tx.Save(product).Error; err != nil {
			return fmt.Errorf("failed to update product: %w", err)
		}

		return nil
	})
}

// BulkUpsert создает или обновляет изделия пачкой
func (r *ProductRepository) BulkUpsert(products []models.Product) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "pn"},
			{Name: "system_id"}, // Составной уникальный ключ
		},
		DoUpdates: clause.AssignmentColumns([]string{"name", "location"}),
	}).Create(&products).Error
}
func (r *ProductRepository) CheckConstraints() error {
	var count int
	if err := r.db.Raw(`
        SELECT COUNT(*) 
        FROM information_schema.table_constraints 
        WHERE table_name = 'products' 
        AND constraint_name = 'uc_products_pn_system'
    `).Scan(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		return errors.New("unique constraint uc_products_pn_system does not exist")
	}
	return nil
}

// repository/product_repository.go

func (r *ProductRepository) Upsert(product *models.Product) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "name"},
			{Name: "system_id"},
		}, // Конфликт по имени
		DoUpdates: clause.AssignmentColumns([]string{"system_id"}), // Обновляем только system_id
	}).Create(product).Error
}

// GetByName возвращает изделие по имени
func (r *ProductRepository) GetByName(name string) (*models.Product, error) {
	var product models.Product
	if err := r.db.Where("name = ?", name).First(&product).Error; err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	return &product, nil
}
func (r *ProductRepository) GetOrCreate(product *models.Product) error {
	return r.db.Where(models.Product{Name: product.Name}).FirstOrCreate(product).Error
}

// LinkToSystem связывает изделие с системой
func (r *ProductRepository) LinkToSystem(product *models.Product, systemID uint) error {
	product.SystemID = &systemID
	return r.db.Model(product).Update("system_id", systemID).Error
}
func (r *ProductRepository) GetWithDetails(id string, product *models.Product) error {
	return r.db.
		Preload("System").
		First(product, id).Error
}
