package models

import (
	"strings"

	"gorm.io/gorm"
)

// FunctionBlock представляет функциональный блок
type FunctionBlock struct {
	gorm.Model
	Tag       string       `gorm:"size:255;not null;uniqueIndex"`
	System    *System      `gorm:"foreignKey:SystemID"`
	CdsType   string       `gorm:"size:50"`
	NodeID    *uint        `gorm:"index"`
	NodeRef   string       `gorm:"size:255"`
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

// ParseFBFromSignal создает/обновляет FunctionBlock из сигнала
func ParseFBFromSignal(signal Signal, direction string) (*FunctionBlock, *FBVariable) {
	fbTag, funcAttr, _ := ParseFBInfo(signal.Tag)
	fb := &FunctionBlock{
		Tag:     fbTag,
		System:  signal.System,
		CdsType: signal.FB,
		Node:    signal.Node,
	}

	variable := &FBVariable{
		SignalTag: signal.Tag,
		FuncAttr:  funcAttr,
		Direction: direction,
	}

	return fb, variable
}
