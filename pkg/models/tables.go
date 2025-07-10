// Package models
// Определение структур для чтения из таблиц, и записи в PostgreSQL
package models

import (
	"time"

	"gorm.io/gorm"
)

type Signal struct {
	gorm.Model

	// Связи

	ProductID *uint    `gorm:"index"` // Опциональная связь с продуктом
	Product   *Product `gorm:"foreignKey:ProductID"`

	NodeID *uint `gorm:"index"` // Опциональная связь с узлом
	Node   *Node `gorm:"foreignKey:NodeID"`

	// Основные поля (из Base)
	Tag         string `gorm:"uniqueIndex;not null"` // Соответствует Base.Tag
	System      string `gorm:"index;size:50"`
	Equipment   string `gorm:"size:100;not null"`
	Name        string `gorm:"size:200;index"`
	Module      string `gorm:"size:100"`
	Channel     string `gorm:"size:50"`
	Crate       string `gorm:"size:50"`
	Place       string `gorm:"index;size:150"`
	Property    string `gorm:"type:TEXT"`
	Address     string `gorm:"size:50"` // Соответствует Base.Adr
	ModbusAddr  string `gorm:"index"`
	NodeRef     string `gorm:"size:255;index"` // Соответствует Base.NodeID
	FB          string `gorm:"size:50"`        // Function Block
	CheckStatus string `gorm:"size:20"`
	Comment     string `gorm:"type:TEXT"`

	// Поля для всех типов сигналов
	SignalType string  `gorm:"size:2;index"` // DI, AI, DO, AO
	Value      float64 `gorm:"type:decimal(20,6)"`

	// Специфичные поля для аналоговых сигналов (AI/AO)
	RangeMin    *float64 `gorm:"type:decimal(20,6)"`
	RangeMax    *float64 `gorm:"type:decimal(20,6)"`
	Unit        *string  `gorm:"size:20"`
	Sign        *string  `gorm:"size:10"`
	WarningLow  *float64 `gorm:"type:decimal(20,6)"` // WL
	WarningHigh *float64 `gorm:"type:decimal(20,6)"` // WH
	AlarmLow    *float64 `gorm:"type:decimal(20,6)"` // AL
	AlarmHigh   *float64 `gorm:"type:decimal(20,6)"` // AH
	Format      *string  `gorm:"size:50"`
	Filter      *string  `gorm:"size:50"`

	// Специфичные поля для дискретных сигналов (DI/DO)
	Category  *string  `gorm:"size:100"`          // DI.Category
	Inversion *string  `gorm:"size:50"`           // DI.Inversion
	TON       *float64 `gorm:"type:decimal(9,3)"` // Timer On Delay
	TOF       *float64 `gorm:"type:decimal(9,3)"` // Timer Off Delay

	// Метаданные
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
type Product struct {
	ID         int    `gorm:"primaryKey" gsheets:"id"`
	PN         string `gorm:"size:100" gsheets:"pn"`
	ProjectPos string `gorm:"size:100" gsheets:"project_pos"`
	Name       string `gorm:"size:200" gsheets:"name"`
	GenPlan    string `gorm:"size:100" gsheets:"gen_plan"`
	Location   string `gorm:"size:200" gsheets:"location"`
}

type Node struct {
	ID   int    `gorm:"primaryKey" gsheets:"id"`
	Main string `gorm:"size:100" gsheets:"main"`
	Sub1 string `gorm:"size:100" gsheets:"sub1"`
	Sub2 string `gorm:"size:100" gsheets:"sub2"`
}

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

type DO struct {
	Base `gsheets:",squash"`
}

type AO struct {
	Base `gsheets:",squash"`
}

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

func (s *Signal) FromAI(ai AI) {
	s.SignalType = "AI"

	// Копируем общие поля из Base
	s.Tag = ai.Tag
	s.System = ai.System
	s.Equipment = ai.Equipment
	s.Name = ai.Name
	s.Module = ai.Module
	s.Channel = ai.Channel
	s.Crate = ai.Crate
	s.Place = ai.Place
	s.Property = ai.Property
	s.Address = ai.Adr
	s.ModbusAddr = ai.ModbusAddr
	s.NodeRef = ai.NodeID
	s.FB = ai.FB
	s.CheckStatus = ai.CheckStatus
	s.Comment = ai.Comment

	// Копируем специфичные поля AI
	s.Value = ai.YMIN // Или другое начальное значение
	s.RangeMin = &ai.YMIN
	s.RangeMax = &ai.YMAX
	s.Unit = &ai.Unit
	s.Sign = &ai.Sign
	s.WarningLow = &ai.WL
	s.WarningHigh = &ai.WH
	s.AlarmLow = &ai.AL
	s.AlarmHigh = &ai.AH
	s.Format = &ai.Format
	s.Filter = &ai.Filter
}

func (s *Signal) FromDI(di DI) {
	s.SignalType = "DI"

	// Копируем общие поля из Base
	s.Tag = di.Tag
	s.System = di.System
	s.Equipment = di.Equipment
	s.Name = di.Name
	s.Module = di.Module
	s.Channel = di.Channel
	s.Crate = di.Crate
	s.Place = di.Place
	s.Property = di.Property
	s.Address = di.Adr
	s.ModbusAddr = di.ModbusAddr
	s.NodeRef = di.NodeID
	s.FB = di.FB
	s.CheckStatus = di.CheckStatus
	s.Comment = di.Comment

	// Копируем специфичные поля DI
	if di.TON > 0 {
		s.TON = &di.TON
	}
	if di.TOF > 0 {
		s.TOF = &di.TOF
	}
	s.Category = &di.Category
	s.Inversion = &di.Inversion

	// Для дискретных сигналов Value = 0 или 1
	// (здесь нужно добавить логику преобразования из DI.Value)
}

// Аналогичные методы FromDO и FromAO
func (s *Signal) FromDO(do DO) {
	s.SignalType = "DO"
	// ... аналогично FromDI ...
}

func (s *Signal) FromAO(ao AO) {
	s.SignalType = "AO"
	// ... аналогично FromAI ...
}

// Кастомный парсер для булевых полей
//func (r *Row) UnmarshalGsheets(fieldName string, value string) error {
//switch fieldName {
//case "Check", "Inversion":
//switch strings.ToLower(value) {
//case "true", "1", "yes", "y", "on", "ok", "pass":
//reflect.ValueOf(r).Elem().FieldByName(fieldName).SetBool(true)
//return nil
//case "false", "0", "no", "n", "off", "fail":
//reflect.ValueOf(r).Elem().FieldByName(fieldName).SetBool(false)
//return nil
//default:
//return fmt.Errorf("invalid boolean value: %s", value)
//}
//}
//return errors.New("unknown field for custom parsing")
//}
