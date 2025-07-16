package models

import (
	"time"

	"gorm.io/gorm"
)

// System представляет систему
type System struct {
	ID             uint            `gorm:"primaryKey"`
	Name           string          `gorm:"size:255;not null"`
	ProjectID      uint            `gorm:"not null"`
	Nodes          []*Node         `gorm:"many2many:node_systems;"`
	Products       []Product       `gorm:"foreignKey:SystemID"`
	FunctionBlocks []FunctionBlock `gorm:"foreignKey:SystemID"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

// Project представляет проект
type Project struct {
	gorm.Model
	Name        string   `gorm:"size:255;not null"`
	Description string   `gorm:"type:TEXT"`
	Systems     []System `gorm:"foreignKey:ProjectID"`
}
