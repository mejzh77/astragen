//Package models
//Определение структур для чтения из таблиц, и записи в PostgreSQL
package models 

import (
	"time"
)

type Product struct {
	ID         int
	PN         string
	ProjectPos string
	Name       string
	GenPlan    string
	Location   string
}
type Base struct {
	//ExternalID  string    `gorm:"uniqueIndex;not null"` // ID из таблицы
	Tag         string    `gsheets:"id"        gorm:"column:uniqueIndex"` // ID из Google-таблицы
	System      string    `gsheets:"system" gorm:"index;size:50"`
	Equipment   string    `gsheets:"equipment" gorm:"size:100;not null"`
	Name        string    `gsheets:"name" gorm:"size:200;index"`
	Product     string    `gsheets:"product" gorm:"size:100"`
	CheckStatus string    `gsheets:"check" gorm:"size:20"`
	FB          string    `gsheets:"fb" gorm:"size:50"` // Function Block
	Comment     string    `gsheets:"comment" gorm:"type:TEXT"`
	Module       string    `gsheets:"module"    gorm:"size:100"`
	Channel      string    `gsheets:"channel"   gorm:"size:50"`
	Crate        string    `gsheets:"crate"     gorm:"size:50"`
	Place        string    `gsheets:"place"     gorm:"index;size:150"`
	Property     string    `gsheets:"property"  gorm:"type:TEXT"`
	Adr      string  `gsheets:"adr" gorm:"size:50"`
	ModbusAddr  string       `gsheets:"modbus" gorm:"index"`
	NodeID      string    `gsheets:"node" gorm:"size:50;index"`
	RecordType  string    `gorm:"size:10"` // 'AI', 'AQ', 'ITF'
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}
type DI struct {
	//gorm.Model
	Base `gsheets:",squash" gorm:"embedded"`

	Category     string    `gsheets:"cat"       gorm:"size:100"`
	Inversion    string    `gsheets:"inversion" gorm:"size:50"`
	TON          float64   `gsheets:"ton"       gorm:"type:decimal(9,3)"` // Timer On Delay (сек)
	TOF          float64   `gsheets:"tof"       gorm:"type:decimal(9,3)"` // Timer Off Delay (сек)
	Comment      string    `gsheets:"comment"   gorm:"type:TEXT"`
	NodeID       string    `gsheets:"node"      gorm:"column:node_id;size:50"`
	LastSyncTime time.Time // Время последней синхронизации
}

type AI struct {
	Base `gsheets:",squash" gorm:"embedded"`
	
	YMIN     float64 `gorm:"type:decimal(12,4)"`
	YMAX     float64 `gorm:"type:decimal(12,4)"`
	Unit     string  `gorm:"size:20"`
	Sign     string  `gorm:"size:10"`
	WL       float64 `gorm:"type:decimal(12,4)"` // Warning Low
	WH       float64 `gorm:"type:decimal(12,4)"` // Warning High
	AL       float64 `gorm:"type:decimal(12,4)"` // Alarm Low
	AH       float64 `gorm:"type:decimal(12,4)"` // Alarm High
	Format   string  `gorm:"size:50"`
	Filter   string  `gorm:"size:50"`
}

type DO struct {
	//gorm.Model
	Base `gsheets:",squash" gorm:"embedded"`

	LastSyncTime time.Time `gorm:"autoUpdateTime"` // Автоматическое обновление
}

type AO struct {
	Base `gsheets:",squash" gorm:"embedded"`

	LastSyncTime time.Time `gorm:"autoUpdateTime"` // Автоматическое обновление
}

type ITF struct {
	Base `gorm:"embedded"`
	
	Lcs      string `gsheets:"acs" gorm:"size:50"`     // Access system
	Protocol string `gsheets:"protocol" gorm:"size:50"`     // Communication protocol
	Address  string `gsheets:"address" gorm:"size:100"`    // Network address
	FuncCode     string `gsheets:"func" gorm:"size:50"`     // Function code
	Offset   int    `gsheets:"offset" gorm:"type:integer"`
	Length   int    `gsheets:"length" gorm:"type:integer"`
	Swap     string `gsheets:"swap" gorm:"size:20"`     // Byte swap
	DataType string `gsheets:"dataType" gorm:"size:50"`     // Data type (int, float, etc.)
	RW       string `gsheets:"rw" gorm:"size:10"`     // Read/Write permission
	Field    string `gsheets:"field" gorm:"size:100"`    // Field name
	Value    string `gsheets:"value" gorm:"type:TEXT"`   // Default value
	Template string `gsheets:"template" gorm:"type:TEXT"`   // Data template
}

type Signal interface {
	DI | DO | AI | AO | ITF
}

type Node struct {
	ID   int
	Main string
	Sub1 string
	Sub2 string
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
