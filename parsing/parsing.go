package parsing

import (
	"encoding/csv"
	"encoding/xml"
	"os"

	"github.com/gocarina/gocsv"
)

// System представляет систему с интерфейсами и переменными
type System struct {
	Name       string
	Itfs       []ITF
	Variables  []Variable
	MbXml      string
	FuncBlocks map[string]config.FB
}

// Variable описывает переменную для Codesys
type Variable struct {
	ID      string
	CdsType string
	Mb      string
	Bit     string
	Pos     string
	Value   string
	Comment string
	Output  bool
}

// ITF описывает интерфейс из CSV файла
type ITF struct {
	ID       string `csv:"id"`
	System   string `csv:"system"`
	Product  string `csv:"product"`
	ACS      string `csv:"acs"`
	Protocol string `csv:"protocol"`
	Address  string `csv:"address"`
	Func     string `csv:"func"`
	Offset   string `csv:"offset"`
	Length   string `csv:"length"`
	Swap     gdBool `csv:"swap"`
	DataType string `csv:"dataType"`
	RW       string `csv:"rw"`
	FB       string `csv:"fb"`
	Field    string `csv:"field"`
	Value    string `csv:"value"`
	Template string `csv:"template"`
	Comment  string `csv:"comment"`
	Check    gdBool `csv:"check"`
}

// ParseCSV парсит CSV файл в список интерфейсов
func ParseCSV(filename string) ([]ITF, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var itfs []ITF
	reader, err := gocsv.NewUnmarshaller(csv.NewReader(file), ITF{})
	if err != nil {
		return nil, err
	}

	for {
		record, err := reader.Read()
		if err != nil {
			break
		}
		itf := record.(ITF)
		if itf.ID == "" || itf.ID == "id" {
			continue
		}
		itfs = append(itfs, itf)
	}

	return itfs, nil
}

// GroupInterfacesBySystem группирует интерфейсы по системам
func GroupInterfacesBySystem(itfs []ITF) (map[string]System, error) {
	systems := make(map[string]System)
	for _, itf := range itfs {
		if system, ok := systems[itf.System]; ok {
			system.Itfs = append(system.Itfs, itf)
			systems[itf.System] = system
		} else {
			systems[itf.System] = System{
				Name: itf.System,
				Itfs: []ITF{itf},
			}
		}
	}
	return systems, nil
}

// ToXML преобразует структуру в XML
func ToXML(data interface{}) ([]byte, error) {
	return xml.MarshalIndent(data, "", "    ")
}
