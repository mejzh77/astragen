package models

import "gorm.io/gorm"

// Product представляет изделие
type Product struct {
	gorm.Model
	PN         string   `gorm:"size:100;uniqueIndex:idx_product_pn_system" gsheets:"pn"`
	ProjectPos string   `gorm:"size:100" gsheets:"project_pos"`
	Name       string   `gorm:"size:200" gsheets:"name"`
	GenPlan    string   `gorm:"size:100" gsheets:"gen_plan"`
	Location   string   `gorm:"size:200" gsheets:"location"`
	SystemID   *uint    `gorm:"index:idx_product_pn_system"`
	System     *System  `gorm:"foreignKey:SystemID"`
	Signals    []Signal `gorm:"foreignKey:ProductID"`
}
