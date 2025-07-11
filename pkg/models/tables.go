// Package models
// Определение структур для чтения из таблиц, и записи в PostgreSQL
package models

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Product представляет изделие
type Product struct {
	gorm.Model
	PN         string   `gorm:"size:100"`
	ProjectPos string   `gorm:"size:100"`
	Name       string   `gorm:"size:200"`
	GenPlan    string   `gorm:"size:100"`
	Location   string   `gorm:"size:200"`
	SystemID   *uint    `gorm:"index"`
	System     *System  `gorm:"foreignKey:SystemID"`
	Signals    []Signal `gorm:"foreignKey:ProductID"`
}

// Node представляет узел системы
type Node struct {
	gorm.Model
	Name           string          `gorm:"size:255"`
	Tag            string          `gorm:"size:100"`
	SystemID       *uint           `gorm:"index"`
	System         *System         `gorm:"foreignKey:SystemID"`
	FunctionBlocks []FunctionBlock `gorm:"foreignKey:NodeID"`
}

// System представляет систему
type System struct {
	ID             uint   `gorm:"primaryKey"`
	Name           string `gorm:"size:255;not null"`
	ProjectID      uint   `gorm:"not null"`
	Nodes          []Node
	Products       []Product
	FunctionBlocks []FunctionBlock `gorm:"foreignKey:SystemID"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

// Project представляет проект
type Project struct {
	gorm.Model
	Name        string   `gorm:"size:255;not null"`
	Description string   `gorm:"type:TEXT"`
	Systems     []System `gorm:"foreignKey:ProjectID"`
}

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
	// Копируем общие поля из Base
	s.Tag = do.Tag
	s.System = do.System
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
	s.System = ao.System
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

// FunctionBlock представляет функциональный блок
type FunctionBlock struct {
	gorm.Model
	Tag       string       `gorm:"size:255;not null;uniqueIndex"`
	System    string       `gorm:"size:100"`
	CdsType   string       `gorm:"size:50"`
	NodeID    *uint        `gorm:"index"`
	SystemID  *uint        `gorm:"index"`
	Node      *Node        `gorm:"foreignKey:NodeID"`
	Variables []FBVariable `gorm:"foreignKey:FBID"`
}

type FBVariable struct {
	gorm.Model
	FBID      uint   `gorm:"index"`
	Direction string `gorm:"size:10;check:direction IN ('input', 'output')"`
	Signal    Signal `gorm:"foreignKey:SignalTag;references:Tag"`
	SignalTag string `gorm:"size:255;not null"`
	FuncAttr  string `gorm:"size:100;not null"` // Часть после последнего '_' в Tag сигнала
}

// ParseFBInfo разбирает тэг сигнала на имя FB и атрибут
func ParseFBInfo(signalTag string) (fbTag, funcAttr string, ok bool) {
	parts := strings.Split(signalTag, "_")
	if len(parts) < 2 {
		return "", "", false
	}
	return strings.Join(parts[:len(parts)-1], "_"), parts[len(parts)-1], true
}
func (p *Project) ToDetailedAPI() gin.H {
	return gin.H{
		"id":        p.ID,
		"name":      p.Name,
		"type":      "project",
		"systems":   p.SystemsToDetailedAPI(),
		"createdAt": p.CreatedAt,
		"updatedAt": p.UpdatedAt,
	}
}

func (p *Project) SystemsToDetailedAPI() []gin.H {
	var systems []gin.H
	for _, s := range p.Systems {
		systems = append(systems, s.ToDetailedAPI())
	}
	return systems
}
func (s *System) ToDetailedAPI() gin.H {
	return gin.H{
		"id":        s.ID,
		"name":      s.Name,
		"type":      "system",
		"projectId": s.ProjectID,
		"nodes":     s.NodesToDetailedAPI(),
		"products":  s.ProductsToDetailedAPI(),
		"createdAt": s.CreatedAt,
		"updatedAt": s.UpdatedAt,
	}
}

func (s *System) NodesToDetailedAPI() []gin.H {
	var nodes []gin.H
	for _, n := range s.Nodes {
		nodes = append(nodes, n.ToDetailedAPI())
	}
	return nodes
}

// Для System
func (s *System) ProductsToDetailedAPI() []gin.H {
	var products []gin.H
	for _, p := range s.Products {
		products = append(products, gin.H{
			"id":        p.ID,
			"name":      p.Name,
			"systemId":  p.SystemID,
			"createdAt": p.CreatedAt,
			"updatedAt": p.UpdatedAt,
		})
	}
	return products
}

func (s *System) FunctionBlocksToDetailedAPI() []gin.H {
	var fbs []gin.H
	for _, fb := range s.FunctionBlocks {
		fbs = append(fbs, gin.H{
			"id":        fb.ID,
			"tag":       fb.Tag,
			"system":    fb.System,
			"cdsType":   fb.CdsType,
			"createdAt": fb.CreatedAt,
			"updatedAt": fb.UpdatedAt,
			"variables": fb.VariablesToDetailedAPI(),
		})
	}
	return fbs
}

// Для Node
func (n *Node) ToDetailedAPI() gin.H {
	return gin.H{
		"id":        n.ID,
		"name":      n.Name,
		"systemId":  n.SystemID,
		"createdAt": n.CreatedAt,
		"updatedAt": n.UpdatedAt,
	}
}

// Для Product
func (p *Product) ToDetailedAPI() gin.H {
	return gin.H{
		"id":        p.ID,
		"name":      p.Name,
		"systemId":  p.SystemID,
		"createdAt": p.CreatedAt,
		"updatedAt": p.UpdatedAt,
	}
}

// Для FunctionBlock
func (fb *FunctionBlock) ToDetailedAPI() gin.H {
	return gin.H{
		"id":        fb.ID,
		"tag":       fb.Tag,
		"system":    fb.System,
		"cdsType":   fb.CdsType,
		"createdAt": fb.CreatedAt,
		"updatedAt": fb.UpdatedAt,
		"variables": fb.VariablesToDetailedAPI(),
	}
}

func (fb *FunctionBlock) VariablesToDetailedAPI() []gin.H {
	var vars []gin.H
	for _, v := range fb.Variables {
		vars = append(vars, gin.H{
			"id":        v.ID,
			"direction": v.Direction,
			"signalTag": v.SignalTag,
			"funcAttr":  v.FuncAttr,
			"fbId":      v.FBID,
			"createdAt": v.CreatedAt,
			"updatedAt": v.UpdatedAt,
		})
	}
	return vars
}

// ParseFBFromSignal создает/обновляет FunctionBlock из сигнала
func ParseFBFromSignal(signal Signal, direction string) (*FunctionBlock, *FBVariable) {
	fbTag, funcAttr, _ := ParseFBInfo(signal.Tag)
	fb := &FunctionBlock{
		Tag:     fbTag,
		System:  signal.System,
		CdsType: signal.FB,
	}

	variable := &FBVariable{
		SignalTag: signal.Tag,
		FuncAttr:  funcAttr,
		Direction: direction,
	}

	return fb, variable
}

// pkg/models/project.go
func (p *Project) ToAPI() gin.H {
	return gin.H{
		"id":      p.ID,
		"name":    p.Name,
		"systems": p.SystemsToAPI(),
	}
}

func (p *Project) SystemsToAPI() []gin.H {
	var systems []gin.H
	for _, s := range p.Systems {
		systems = append(systems, s.ToAPI())
	}
	return systems
}
func (s *System) ToAPI() gin.H {
	return gin.H{
		"id":             s.ID,
		"name":           s.Name,
		"projectId":      s.ProjectID,
		"nodes":          s.NodesToAPI(),
		"products":       s.ProductsToAPI(),
		"functionBlocks": s.FunctionBlocksToAPI(),
	}
}

func (s *System) ProductsToAPI() []gin.H {
	var products []gin.H
	for _, p := range s.Products {
		products = append(products, gin.H{
			"id":        p.ID,
			"name":      p.Name,
			"systemId":  p.SystemID,
			"createdAt": p.CreatedAt,
		})
	}
	return products
}

func (s *System) FunctionBlocksToAPI() []gin.H {
	var fbs []gin.H
	for _, fb := range s.FunctionBlocks {
		fbs = append(fbs, gin.H{
			"id":        fb.ID,
			"tag":       fb.Tag,
			"system":    fb.System,
			"cdsType":   fb.CdsType,
			"variables": fb.VariablesToAPI(),
		})
	}
	return fbs
}

func (fb *FunctionBlock) VariablesToAPI() []gin.H {
	var vars []gin.H
	for _, v := range fb.Variables {
		vars = append(vars, gin.H{
			"id":        v.ID,
			"direction": v.Direction,
			"signalTag": v.SignalTag,
			"funcAttr":  v.FuncAttr,
			"fbId":      v.FBID,
		})
	}
	return vars
}
func (s *System) NodesToAPI() []gin.H {
	var nodes []gin.H
	for _, n := range s.Nodes {
		nodes = append(nodes, n.ToAPI())
	}
	return nodes
}
func (n *Node) ToAPI() gin.H {
	return gin.H{
		"id":       n.ID,
		"name":     n.Name,
		"systemId": n.SystemID,
	}
}
