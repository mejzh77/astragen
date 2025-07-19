package models

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FunctionBlock представляет функциональный блок
type FunctionBlock struct {
	gorm.Model
	Tag       string       `gorm:"size:255;not null;uniqueIndex"`
	System    *System      `gorm:"foreignKey:SystemID"`
	Call      string       `gorm:"type:TEXT"`
	OMX       string       `gorm:"type:TEXT"`
	OPC       string       `gorm:"type:TEXT"`
	CdsType   string       `gorm:"size:50"`
	Equipment string       `gorm:"size:50"`
	NodeID    *uint        `gorm:"index"`
	NodeRef   string       `gorm:"size:255"`
	SystemID  *uint        `gorm:"index"`
	Node      *Node        `gorm:"foreignKey:NodeID"`
	Variables []FBVariable `gorm:"foreignKey:FBID"`
}

type FBCallParams struct {
	Tag     string
	CdsType string
	In      IOPair
	Out     IOPair
}

type IOPair map[string]string

type OPCItem struct {
	Binding    string
	NodePath   string
	Namespace  string
	NodeIdType string
	NodeId     string
}

type OPCTemplate struct {
	Binding    string
	NodePath   string
	Namespace  string
	BasePath   string
	NodePrefix string
	PathSuffix []string
}

type OMXData struct {
	FB   *FunctionBlock
	UUID string
}

func (fb *FunctionBlock) GenerateSTCode(fbTemplate string, defaultInputs, defaultOutputs map[string]string) (string, error) {
	// Разделяем переменные на входы и выходы
	inputs := make(IOPair)
	outputs := make(IOPair)
	for _, v := range fb.Variables {
		switch v.Direction {
		case "input":
			inputs[defaultInputs[v.FuncAttr]] = v.SignalTag
		case "output":
			outputs[defaultOutputs[v.FuncAttr]] = v.SignalTag
		}
	}

	// Подготавливаем данные для шаблона
	data := FBCallParams{
		Tag:     fb.Tag,
		CdsType: fb.CdsType,
		In:      inputs,
		Out:     outputs,
	}

	// Создаем и выполняем шаблон
	tmpl, err := template.New("fbCall").Parse(fbTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (fb *FunctionBlock) GenerateOMX(tmplBase string, attributes map[string]string) (string, error) {
	// Шаг 2: Генерируем блок атрибутов
	tmplWithAttrs := strings.Replace(tmplBase, `</object>`, "", 1)
	for k, v := range attributes {
		tmplWithAttrs += fmt.Sprintf("<attribute type=\"%s\" value=\"{{%s}}\"></attribute>\n", k, v)
	}
	tmplWithAttrs += `</object>`
	tmpl, err := template.New("attr").Parse(tmplWithAttrs)
	if err != nil {
		return "", err
	}
	data := OMXData{
		FB:   fb,
		UUID: uuid.NewString(),
	}
	// Шаг 4: Применяем основной шаблон
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (fb *FunctionBlock) GenerateOPC(mapping OPCTemplate) (string, error) {
	data := struct {
		FB   *FunctionBlock
		UUID string
	}{
		FB:   fb,
		UUID: uuid.NewString(),
	}

	var items []OPCItem
	for _, itemDef := range mapping.PathSuffix {
		// Генерируем пути
		nodePath, err := executeTemplate(mapping.BasePath, data)
		if err != nil {
			return "", err
		}
		nodePath += "." + itemDef

		nodeId, err := executeTemplate(mapping.NodePrefix, data)
		if err != nil {
			return "", err
		}
		nodeId += "." + itemDef

		items = append(items, OPCItem{
			Binding:    mapping.Binding,
			NodePath:   nodePath,
			Namespace:  mapping.Namespace,
			NodeIdType: "string",
			NodeId:     nodeId,
		})
	}

	// Генерация XML
	var buf bytes.Buffer
	buf.WriteString(`<opc-import xmlns="urn:prosoft:opc-import">` + "\n")

	for _, item := range items {
		buf.WriteString(`  <item Binding="` + item.Binding + `">` + "\n")
		buf.WriteString(`    <node-path>` + item.NodePath + `</node-path>` + "\n")
		buf.WriteString(`    <namespace>` + item.Namespace + `</namespace>` + "\n")
		buf.WriteString(`    <nodeIdType>` + item.NodeIdType + `</nodeIdType>` + "\n")
		buf.WriteString(`    <nodeId>` + item.NodeId + `</nodeId>` + "\n")
		buf.WriteString(`  </item>` + "\n")
	}

	buf.WriteString(`</opc-import>`)
	return buf.String(), nil
}
func executeTemplate(tmplStr string, data interface{}) (string, error) {
	tmpl, err := template.New("").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
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
		Tag:       fbTag,
		System:    signal.System,
		CdsType:   signal.FB,
		NodeID:    signal.NodeID,
		Equipment: signal.Equipment,
	}

	variable := &FBVariable{
		SignalTag: signal.Tag,
		FuncAttr:  funcAttr,
		Direction: direction,
	}

	return fb, variable
}
