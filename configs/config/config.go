package config

import "github.com/mejzh77/astragen/pkg/models"

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string // "disable" или "require"
}
type SheetConfig struct {
	SheetName  string
	SignalType string      // "DI", "AI", "DO", "AO"
	Model      interface{} // Указатель на соответствующую модель (например, &models.DI{})
}

var AppConfig = struct {
	DB            DatabaseConfig
	SpreadsheetID string
	Sheets        []SheetConfig
}{
	DB: DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "user",
		Password: "advengauser",
		Name:     "astragen",
		SSLMode:  "disable",
	},
	SpreadsheetID: "1GAUwJRTtrBT4gr1y3ETsCSlHojrc7VCD2GlGDUM53kQ",
	Sheets: []SheetConfig{
		{
			SheetName:  "DI",
			SignalType: "DI",
			Model:      &models.DI{},
		},
		{
			SheetName:  "AI",
			SignalType: "AI",
			Model:      &models.AI{},
		},
		{
			SheetName:  "DQ",
			SignalType: "DO",
			Model:      &models.DO{},
		},
		{
			SheetName:  "AQ",
			SignalType: "AO",
			Model:      &models.AO{},
		},
	},
}
