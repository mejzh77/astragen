package models

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FunctionBlock представляет функциональный блок
type FunctionBlock struct {
	gorm.Model
	Tag         string       `gorm:"size:255;not null;uniqueIndex"`
	System      *System      `gorm:"foreignKey:SystemID"`
	Declaration string       `gorm:"type:TEXT"`
	Call        string       `gorm:"type:TEXT"`
	OMX         string       `gorm:"type:TEXT"`
	OPC         string       `gorm:"type:TEXT"`
	CdsType     string       `gorm:"size:50"`
	Primary     bool         `gorm:"not null;default:false"`
	Equipment   string       `gorm:"size:50"`
	NodeID      *uint        `gorm:"index"`
	NodeRef     string       `gorm:"size:255"`
	SystemID    *uint        `gorm:"index"`
	Address     string       `gorm:"size:50"`
	Node        *Node        `gorm:"foreignKey:NodeID"`
	Variables   []FBVariable `gorm:"foreignKey:FBID"`
	Description string       `gorm:"type:TEXT"`
	Name        string       `gorm:"size:255"`
	Comment     string       `gorm:"type:TEXT"`
}

type FBCallParams struct {
	Tag     string
	CdsType string
	Address string
	Comment string
	Node    string
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

func ProcessIOPair(pairIn, pairOut map[string]string, fb *FunctionBlock) (IOPair, IOPair) {
	inputs := make(IOPair)
	outputs := make(IOPair)
	if fb.CdsType == "DI" {
		fmt.Print("")
	}
	for lhs, rhs := range pairIn {
		if rhs == "address" {
			inputs[rhs] = lhs
		}
	}
	for lhs, rhs := range pairOut {
		if rhs == "address" {
			outputs[rhs] = lhs
		}
	}
	for _, v := range fb.Variables {
		switch v.Direction {
		case "input":
			for lhs, rhs := range pairIn {
				if rhs == "address" {
					inputs[rhs] = lhs
				}
				parts := strings.Split(rhs, ".")
				if v.FuncAttr == parts[0] {
					if len(parts) > 1 {
						inputs[lhs] = v.CdsType + "." + v.SignalTag + "." + parts[1]
					} else {
						inputs[lhs] = v.CdsType + "." + v.SignalTag
					}
				}
			}
		case "output":
			for lhs, rhs := range pairOut {
				if rhs == "address" {
					outputs[rhs] = lhs
				}
				parts := strings.Split(rhs, ".")
				if v.FuncAttr == parts[0] {
					if len(parts) > 1 {
						outputs[lhs] = v.CdsType + "." + v.SignalTag + parts[1]
					} else {
						outputs[lhs] = v.CdsType + "." + v.SignalTag
					}
				}
			}
		}
	}
	return inputs, outputs
}
func (fb *FunctionBlock) GenerateSTCode(fbTemplate string, defaultInputs, defaultOutputs map[string]string) (string, error) {
	// Разделяем переменные на входы и выходы
	inputs, outputs := ProcessIOPair(defaultInputs, defaultOutputs, fb)

	var nodeName string
	if fb.Node != nil {
		nodeName = fb.Node.Name
	}
	// Подготавливаем данные для шаблона
	data := FBCallParams{
		Tag:     fb.Tag,
		CdsType: fb.CdsType,
		Address: fb.Address,
		Comment: fb.Comment,
		Node:    nodeName,
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

func FormatVarDeclaration(tag, varType string, indentSize, totalWidth int) string {
	// Создаем отступ
	indent := strings.Repeat(" ", indentSize)

	// Формируем строку с выравниванием
	// Используем fmt.Sprintf с минимальной шириной для первой части
	formatted := fmt.Sprintf("%s%-*s %s;",
		indent,
		totalWidth-indentSize-len(" ")-len(";"),
		tag+":",
		varType)

	return formatted
}
func (fb *FunctionBlock) GenerateSTDecl() (string, error) {
	return FormatVarDeclaration(fb.Tag, "FB_"+fb.CdsType, 4, 40), nil
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
func executeTemplate(tmplStr string, data interface{}, funcs ...template.FuncMap) (string, error) {
	tmpl := template.New("")
	if len(funcs) > 0 {
		tmpl = tmpl.Funcs(funcs[0])
	}
	tmpl, err := tmpl.Parse(tmplStr)
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
	CdsType   string `gorm:"size:30"`
	Signal    Signal `gorm:"foreignKey:SignalTag;references:Tag"`
	Address   string `gorm:"size:255"`
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
func decrementString(channel string) (string, error) {
	num, err := strconv.Atoi(channel)
	if err != nil {
		return "", fmt.Errorf("invalid channel number: %v", err)
	}
	return strconv.Itoa(num - 1), nil
}
func formatNumber(input string, length int) string {
	if len(input) >= length {
		return input
	}

	// Дополняем нулями слева
	return strings.Repeat("0", length-len(input)) + input
}

// ParseFBFromSignal создает/обновляет FunctionBlock из сигнала
func ParseFBFromSignal(signal Signal, direction string, addressTmpl string) (*FunctionBlock, *FBVariable, error) {
	fbTag, funcAttr, _ := ParseFBInfo(signal.Tag)
	addr, err := executeTemplate(addressTmpl, signal, template.FuncMap{"decrement": decrementString, "format_number": formatNumber})
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка выполнения шаблона для сигнала %s: %v", signal.Tag, err)
	}
	fb := &FunctionBlock{
		Tag:       fbTag,
		System:    signal.System,
		SystemID:  signal.SystemID,
		CdsType:   signal.FB,
		NodeID:    signal.NodeID,
		NodeRef:   signal.NodeRef,
		Equipment: signal.Equipment,
	}

	variable := &FBVariable{
		SignalTag: signal.Tag,
		Address:   addr,
		CdsType:   signal.SignalType,
		FuncAttr:  funcAttr,
		Direction: direction,
	}

	return fb, variable, nil
}

func UpdateAddress(signal Signal, addressTmpl string) (string, error) {
	addr, err := executeTemplate(addressTmpl, signal, template.FuncMap{"decrement": decrementString, "format_number": formatNumber})
	if err != nil {
		return signal.Address, fmt.Errorf("ошибка выполнения шаблона для сигнала %s: %v", signal.Tag, err)
	} else {
		return addr, nil
	}
}

func ParseFromSignal(signal Signal, addressTmpl string) (*FunctionBlock, error) {
	addr, err := executeTemplate(addressTmpl, signal, template.FuncMap{"decrement": decrementString, "format_number": formatNumber})
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения шаблона для сигнала %s: %v", signal.Tag, err)
	}
	fb := &FunctionBlock{
		Tag:       signal.Tag,
		System:    signal.System,
		SystemID:  signal.SystemID,
		Primary:   true,
		CdsType:   signal.FB,
		NodeID:    signal.NodeID,
		NodeRef:   signal.NodeRef,
		Equipment: signal.Equipment,
		Address:   addr,
		Name:      signal.Name,
		Comment: fmt.Sprintf("%s\n%s\nproduct: %s; crate: %s; module: %s; channel: %s",
			signal.NodeRef, signal.Name, signal.ProductRef, signal.Crate, signal.Module, signal.Channel),
	}

	return fb, nil
}
