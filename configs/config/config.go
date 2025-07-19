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

type OMXConfig struct {
	Template   string            `yaml:"template"`
	Attributes map[string]string `yaml:"attributes"`
}

type FBConfig struct {
	Template string            `yaml:"st_template"`
	In       map[string]string `yaml:"in"`
	Out      map[string]string `yaml:"out"`
	OMX      OMXConfig         `yaml:"omx"`
	OPC      OPCConfig         `yaml:"opc"`
}

type OPCConfig struct {
	Items []string `yaml:"items"`
}

type OPCItemTemplate struct {
	BasePath   string `yaml:"base_path"`
	NodePrefix string `yaml:"node_prefix"`
	Namespace  string `yaml:"namespace"`
	NodeIdType string `yaml:"nodeIdType"`
	Binding    string `yaml:"binding"`
}

type AppConfig struct {
	DB             DatabaseConfig      `yaml:"db"`
	SpreadsheetID  string              `yaml:"spreadsheet_id"`
	Update         bool                `yaml:"update"`
	Sheets         []SheetConfig       `yaml:"sheets"`
	FunctionBlocks map[string]FBConfig `yaml:"function_blocks"`
	Systems        []string            `yaml:"systems"`
	NodeSheet      string              `yaml:"nodesheet"`
	DefaultOPCItem OPCItemTemplate     `yaml:"default_opc"`
	ProductSheet   string              `yaml:"productsheet"`
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
