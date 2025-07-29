package models

// Base и другие структуры остаются как у вас, только убираем gorm теги:
type Base struct {
	Tag         string `gsheets:"id"`
	System      string `gsheets:"system"`
	Equipment   string `gsheets:"equipment"`
	Name        string `gsheets:"name"`
	Product     string `gsheets:"product"`
	CheckStatus string `gsheets:"check"`
	FB          string `gsheets:"fb"`
	Comment     string `gsheets:"comment"`
	Module      string `gsheets:"module"`
	Channel     string `gsheets:"channel"`
	Crate       string `gsheets:"crate"`
	Place       string `gsheets:"place"`
	Property    string `gsheets:"property"`
	Adr         string `gsheets:"adr"`
	ModbusAddr  string `gsheets:"modbus"`
	NodeID      string `gsheets:"node"`
}

type DI struct {
	Base `gsheets:",squash"`

	Category  string  `gsheets:"cat"`
	Inversion string  `gsheets:"inversion"`
	TON       float64 `gsheets:"ton"`
	TOF       float64 `gsheets:"tof"`
}

type AI struct {
	Base `gsheets:",squash"`

	YMIN   float64 `gsheets:"YMIN"`
	YMAX   float64 `gsheets:"YMAX"`
	Unit   string  `gsheets:"unit"`
	Sign   string  `gsheets:"sign"`
	WL     float64 `gsheets:"WL"`
	WH     float64 `gsheets:"WH"`
	AL     float64 `gsheets:"AL"`
	AH     float64 `gsheets:"AH"`
	Format string  `gsheets:"format"`
	Filter string  `gsheets:"filter"`
}

type DQ struct {
	Base `gsheets:",squash"`
}

type AQ struct {
	Base `gsheets:",squash"`
}

// Модель для загрузки узлов из Google Sheets
type SheetNode struct {
	Name    string `gsheets:"Обозначение"`
	Tag     string `gsheets:"Тэг"`
	Systems string `gsheets:"Система"`
	// Другие поля по необходимости
}

// Модель для загрузки изделий из Google Sheets
type SheetProduct struct {
	PN       string `gsheets:"Заводской номер"`
	Name     string `gsheets:"Проектная позиция"`
	System   string `gsheets:"Система"`
	Location string `gsheets:"Название размещения"`
	Tag      string `gsheets:"tag"`
	// Другие поля по необходимости
}

type SheetFB struct {
	System      string `gsheets:"system"`
	CdsType     string `gsheets:"cds_type"`
	Tag         string `gsheets:"tag"`
	Name        string `gsheets:"name"`
	Description string `gsheets:"description"`
}
