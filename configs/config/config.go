package config

import (
	"log"
	"os"

	"github.com/mejzh77/astragen/pkg/models"
	"gopkg.in/yaml.v3"
)

var Cfg *AppConfig

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"ssl_mode"` // "disable" или "require"
}

type SheetConfig struct {
	SheetName  string      `yaml:"sheet_name"`
	SignalType string      `yaml:"signal_type"` // "DI", "AI", "DO", "AO"
	Model      interface{} `yaml:"-"`           // Указатель на модель (не для YAML)
}
type VarsConfig struct {
	In  map[string]string `yaml:"in"`
	Out map[string]string `yaml:"out"`
}

type FBConfig map[string]VarsConfig

type AppConfig struct {
	DB             DatabaseConfig `yaml:"db"`
	SpreadsheetID  string         `yaml:"spreadsheet_id"`
	Update         bool           `yaml:"update"`
	Sheets         []SheetConfig  `yaml:"sheets"`
	FunctionBlocks FBConfig       `yaml:"function_blocks"`
	Systems        []string       `yaml:"systems"`
	NodeSheet      string         `yaml:"nodesheet"`

	ProductSheet string `yaml:"productsheet"`
}

// LoadConfig загружает конфиг из файла
func LoadConfig(path string) *AppConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	// Инициализируем модели для листов
	cfg.Sheets = []SheetConfig{
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
	}

	return &cfg
}
