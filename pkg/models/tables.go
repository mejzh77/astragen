// Package models
// Определение структур для чтения из таблиц, и записи в PostgreSQL
package models

type ITF struct {
	Base `gorm:"embedded"`

	Lcs      string `gsheets:"acs" gorm:"size:50"`      // Access system
	Protocol string `gsheets:"protocol" gorm:"size:50"` // Communication protocol
	Address  string `gsheets:"address" gorm:"size:100"` // Network address
	FuncCode string `gsheets:"func" gorm:"size:50"`     // Function code
	Offset   int    `gsheets:"offset" gorm:"type:integer"`
	Length   int    `gsheets:"length" gorm:"type:integer"`
	Swap     string `gsheets:"swap" gorm:"size:20"`       // Byte swap
	DataType string `gsheets:"dataType" gorm:"size:50"`   // Data type (int, float, etc.)
	RW       string `gsheets:"rw" gorm:"size:10"`         // Read/Write permission
	Field    string `gsheets:"field" gorm:"size:100"`     // Field name
	Value    string `gsheets:"value" gorm:"type:TEXT"`    // Default value
	Template string `gsheets:"template" gorm:"type:TEXT"` // Data template
}

type Cable struct {
	ID                int
	Name              string
	ProjectPosFrom    string
	TerminalGroupFrom string
	TerminalFrom      string
	ProjectPosTo      string
	TerminalGroupTo   string
	TerminalTo        string
	Product           string
	Cable             string
}

type Row interface {
	Product | Signal | Cable | Node
}
