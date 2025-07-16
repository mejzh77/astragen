package models

import "gorm.io/gorm"

// Node представляет узел системы
type Node struct {
	gorm.Model
	Name           string          `gorm:"size:255;uniqueIndex:idx_node_name_system" gsheets:"name"`
	Tag            string          `gorm:"size:100" gsheets:"tag"`
	SystemID       *uint           `gorm:"index:idx_node_name_system"`
	Systems        []*System       `gorm:"many2many:node_systems;"`
	FunctionBlocks []FunctionBlock `gorm:"foreignKey:NodeID"`
}
