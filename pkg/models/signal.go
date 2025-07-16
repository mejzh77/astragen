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
	Tag         string  `gorm:"uniqueIndex;not null"` // Соответствует Base.Tag
	SystemID    *uint   `gorm:"index"`                // Внешний ключ
	System      *System `gorm:"foreignKey:SystemID"`  // Связь
	SystemRef   string  `gorm:"-"`                    // Временное поле для загрузки из Google Sheets
	Equipment   string  `gorm:"size:100;not null"`
	Name        string  `gorm:"size:200;index"`
	Module      string  `gorm:"size:100"`
	Channel     string  `gorm:"size:50"`
	Crate       string  `gorm:"size:50"`
	Place       string  `gorm:"index;size:150"`
	Property    string  `gorm:"type:TEXT"`
	Address     string  `gorm:"size:50"` // Соответствует Base.Adr
	ModbusAddr  string  `gorm:"index"`
	NodeRef     string  `gorm:"size:255;index"` // Соответствует Base.NodeID
	FB          string  `gorm:"size:50"`        // Function Block
	CheckStatus string  `gorm:"size:20"`
	Comment     string  `gorm:"type:TEXT"`

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

func (s *Signal) FromAI(ai AI) {
	s.SignalType = "AI"

	// Копируем общие поля из Base
	s.Tag = ai.Tag
	s.SystemRef = ai.System
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
	s.SystemRef = di.System
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
	// Копируем общие поля из Base
	s.Tag = do.Tag
	s.SystemRef = do.System
	s.Equipment = do.Equipment
	s.Name = do.Name
	s.Module = do.Module
	s.Channel = do.Channel
	s.Crate = do.Crate
	s.Place = do.Place
	s.Property = do.Property
	s.Address = do.Adr
	s.ModbusAddr = do.ModbusAddr
	s.NodeRef = do.NodeID
	s.FB = do.FB
	s.CheckStatus = do.CheckStatus
	s.Comment = do.Comment
}

func (s *Signal) FromAO(ao AO) {
	s.SignalType = "AO"
	s.Tag = ao.Tag
	s.SystemRef = ao.System
	s.Equipment = ao.Equipment
	s.Name = ao.Name
	s.Module = ao.Module
	s.Channel = ao.Channel
	s.Crate = ao.Crate
	s.Place = ao.Place
	s.Property = ao.Property
	s.Address = ao.Adr
	s.ModbusAddr = ao.ModbusAddr
	s.NodeRef = ao.NodeID
	s.FB = ao.FB
	s.CheckStatus = ao.CheckStatus
	s.Comment = ao.Comment
}
