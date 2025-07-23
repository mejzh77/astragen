package models

import "gorm.io/gorm"

// Product представляет изделие
type Product struct {
	gorm.Model
	PN         string   `gorm:"size:100;uniqueIndex:idx_product_pn_system"`
	ProjectPos string   `gorm:"size:100"`
	Name       string   `gorm:"size:200"`
	Tag        string   `gorm:"size:200"`
	GenPlan    string   `gorm:"size:100"`
	Location   string   `gorm:"size:200"`
	SystemID   *uint    `gorm:"index:idx_product_pn_system"`
	System     *System  `gorm:"foreignKey:SystemID"`
	Signals    []Signal `gorm:"foreignKey:ProductID"`
}
