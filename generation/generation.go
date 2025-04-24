package generation

import (
	"fmt"
	"os"
	"text/template"

	"github.com/jzh7/astragen/config"
	"github.com/jzh7/astragen/parsing"
)

// GenerateVariables генерирует объявления переменных для Codesys
func GenerateVariables(sys parsing.System, cfg *config.Config) {
	fmt.Printf("%s\n/////\n", sys.Name)
	for _, v := range sys.Variables {
		fmt.Printf("%s:\t%s;\t//%s\n", v.ID, v.CdsType, v.Comment)
	}
	fmt.Printf("\n///////////\n")
	fmt.Printf("\n///////Чтение переменных из Modbus//////////\n")

	for _, v := range sys.Variables {
		tmplName := v.CdsType
		if v.Value != "" {
			tmplName += "_ref"
		}
		if tmpl, ok := cfg.Templates["r"][tmplName]; ok {
			if !v.Output {
				t := template.Must(template.New(tmplName).Parse(tmpl))
				err := t.Execute(os.Stdout, v)
				if err != nil {
					fmt.Printf("Failed to execute template: %v\n", err)
				}
			}
		}
	}
}

// GenerateFunctionalBlocks генерирует функциональные блоки
func GenerateFunctionalBlocks(sys parsing.System, cfg *config.Config) {
	fmt.Printf("///////////Функциональные блоки//////////\n")
	for _, v := range sys.FuncBlocks {
		fmt.Printf("%s:\t%s;\n", v.ID, v.CdsType)
	}
	for _, v := range sys.FuncBlocks {
		if tmpl, ok := cfg.Templates["fb"][v.Template]; ok {
			t := template.Must(template.New(v.Template).Parse(tmpl))
			err := t.Execute(os.Stdout, v)
			if err != nil {
				fmt.Printf("Failed to execute template: %v\n", err)
			}
		} else {
			t := template.Must(template.New("default").Parse(cfg.Templates["fb"]["default"]))
			err := t.Execute(os.Stdout, v)
			if err != nil {
				fmt.Printf("Failed to execute template: %v\n", err)
			}
		}
	}
}

// GenerateOPC генерирует OPC файл
func GenerateOPC(sys parsing.System, cfg *config.Config) {
	opc := parsing.NewOPCMap()
	for _, fb := range sys.FuncBlocks {
		opc.AddObject(sys.Name, cfg.OPCType[fb.CdsType], fb.ID)
	}
	err := opc.SaveToFile(sys.Name + "_opc.xml")
	if err != nil {
		fmt.Printf("Failed to save OPC file: %v\n", err)
	}
}

// GenerateOMX генерирует OMX файл
func GenerateOMX(sys parsing.System, cfg *config.Config) {
	omx := parsing.NewOMX()
	for _, v := range sys.FuncBlocks {
		ASType := v.CdsType
		if val, ok := cfg.OMXType[v.CdsType]; ok {
			ASType = val
		}
		omx.AddObject(ASType, v.ID)
	}
	xmlText, err := parsing.ToXML(omx)
	if err != nil {
		fmt.Printf("Failed to marshal OMX: %v\n", err)
	}
	err = os.WriteFile(sys.Name+".omx-export", xmlText, 0644)
	if err != nil {
		fmt.Printf("Failed to write OMX file: %v\n", err)
	}
}
