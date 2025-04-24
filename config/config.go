package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// Config представляет структуру конфигурационного файла
type Config struct {
	GSID      string              `yaml:"GoogleID"`
	Update    bool                `yaml:"update"`
	System    string              `yaml:"system"`
	Templates map[string]Template `yaml:"templates"`
	Vars      map[string]FB       `yaml:"vars"`
	OMXType   map[string]string   `yaml:"OMXTypes"`
	OPCType   map[string]string   `yaml:"OPCTypes"`
}

// Template представляет шаблоны для генерации кода
type Template map[string]string

// FB описывает функциональный блок
type FB struct {
	ID       string
	CdsType  string
	Template string
	In       map[string]string
	Out      map[string]string
}

// LoadConfig загружает конфигурацию из YAML файла
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
