package tables

import (
	"fmt"
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

type DI struct {
	gorm.Model
	Tag          string    `gsheets:"id"        gorm:"column:uniqueIndex"` // ID из Google-таблицы
	System       string    `gsheets:"system"    gorm:"index;size:100"`
	Equipment    string    `gsheets:"equipment" gorm:"size:200"`
	Name         string    `gsheets:"name"      gorm:"size:200"`
	Place        string    `gsheets:"place"     gorm:"index;size:150"`
	Product      string    `gsheets:"product"   gorm:"size:150"`
	Module       string    `gsheets:"module"    gorm:"size:100"`
	Channel      string    `gsheets:"channel"   gorm:"size:50"`
	Crate        string    `gsheets:"crate"     gorm:"size:50"`
	CheckStatus  string    `gsheets:"check"     gorm:"column:check_status;size:50"`
	Category     string    `gsheets:"cat"       gorm:"size:100"`
	Property     string    `gsheets:"property"  gorm:"type:TEXT"`
	FB           string    `gsheets:"fb"        gorm:"size:50"` // Function Block
	Inversion    string    `gsheets:"inversion" gorm:"size:50"`
	TON          float64   `gsheets:"ton"       gorm:"type:decimal(9,3)"` // Timer On Delay (сек)
	TOF          float64   `gsheets:"tof"       gorm:"type:decimal(9,3)"` // Timer Off Delay (сек)
	Comment      string    `gsheets:"comment"   gorm:"type:TEXT"`
	ModbusAddr   int       `gsheets:"modbus"    gorm:"column:modbus_addr"`
	NodeID       string    `gsheets:"node"      gorm:"column:node_id;size:50"`
	LastSyncTime time.Time // Время последней синхронизации
}

type AI struct {
	ID        int
	Tag       string
	Name      string
	Product   string
	Module    string
	Channel   string
	Equipment string
	Check     string
	Property  string
	Term      string
	Node      string
}

type DO struct {
	ID        int
	Tag       string
	Name      string
	Product   string
	Module    string
	Channel   string
	Equipment string
	Check     string
	Property  string
	Term      string
	Node      string
}

type AO struct {
	ID        int
	Tag       string
	Name      string
	Product   string
	Module    string
	Channel   string
	Equipment string
	Check     string
	Property  string
	Term      string
	Node      string
}

type ITF struct {
	ID        int
	Tag       string
	Name      string
	Product   string
	Module    string
	Channel   string
	Equipment string
	Check     string
	Property  string
	Term      string
	Node      string
}

type Signal interface {
	DI | DO | AI | AO
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

func main() {
	fmt.Println("vim-go")
}
