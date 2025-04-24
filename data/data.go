package data

import (
	"cmp"
	"fmt"
	"strconv"
	"strings"

	"github.com/jzh7/astragen/gd2csv"
	"github.com/jzh7/astragen/parsing"
	"golang.org/x/exp/slices"
)

// UpdateFromGoogle обновляет данные из Google Sheets
func UpdateFromGoogle(gsid, destDir string) error {
	return gd2csv.Update(gsid, destDir)
}

// ProcessSystem обрабатывает систему, сортирует интерфейсы и генерирует переменные
func ProcessSystem(sys *parsing.System, cfg *config.Config) error {
	// Сортировка интерфейсов по функции и смещению
	slices.SortFunc(sys.Itfs, func(a, b parsing.ITF) int {
		aa, _ := strconv.Atoi(a.Offset)
		bb, _ := strconv.Atoi(b.Offset)
		return strings.Compare(a.Func, b.Func) + cmp.Compare(aa, bb)
	})

	mbXml := parsing.ModbusOuterChannels{
		OuterChannels: make([]parsing.OuterChannel, 0),
	}
	prevFC := sys.Itfs[0].Func
	startPos := 0
	pos := -1
	fbs := make(map[string]config.FB, 0)
	var variables []parsing.Variable

	for i, itf := range sys.Itfs {
		mbName := "arwMB_" + strconv.Itoa(startPos)

		// Обработка полей Modbus
		if itf.Field == "0" {
			if itf.Func != prevFC || (i != 0 && strDelta(itf.Offset, sys.Itfs[i-1].Offset) > 1) {
				mbLen := strconv.Itoa(strDelta(itf.Offset, strconv.Itoa(startPos)))
				ch := parsing.OuterChannel{
					Name:         mbName,
					Description:  "",
					Offset:       strconv.Itoa(startPos),
					Length:       mbLen,
					FunctionCode: fc2Text(prevFC),
					CycleTime:    "1000",
					ChannelType:  "Timer",
				}
				variables = append(variables, parsing.Variable{
					ID:      mbName,
					CdsType: "ARRAY[0.." + mbLen + "] OF WORD",
					Comment: "Массив значений Modbus по адресам " + strconv.Itoa(startPos) + ".." + itf.Offset,
				})
				startPos, _ = strconv.Atoi(itf.Offset)
				mbXml.OuterChannels = append(mbXml.OuterChannels, ch)
				pos = 0
			}
			pos++
		}

		// Создание переменной
		variable := parsing.Variable{
			ID:      strings.TrimSpace(itf.ID),
			CdsType: strings.TrimSpace(itf.DataType),
			Mb:      mbName,
			Bit:     strings.TrimSpace(itf.Field),
			Pos:     strconv.Itoa(pos),
			Value:   strings.TrimSpace(itf.Value),
			Comment: strings.TrimSpace(itf.Comment),
			Output:  itf.RW == "w",
		}

		// Обработка функциональных блоков
		tokens := strings.Split(itf.ID, "_")
		fbName := strings.Join(tokens[0:len(tokens)-1], "_")
		if len(tokens) < 2 {
			fmt.Printf("Ошибка именования переменной %s\n", itf.ID)
			continue
		}
		varToken := tokens[len(tokens)-1]
		if _, ok := fbs[fbName]; !ok {
			fbs[fbName] = config.FB{
				ID:       fbName,
				Template: itf.Template,
				CdsType:  itf.FB,
				In:       make(map[string]string, 0),
				Out:      make(map[string]string, 0),
			}
		}

		// Привязка входов и выходов к функциональным блокам
		for k, v := range cfg.Vars[itf.FB].In {
			if v == varToken {
				fbs[fbName].In[k] = variable.ID
				break
			}
		}
		for k, v := range cfg.Vars[itf.FB].Out {
			if v == varToken {
				fbs[fbName].Out[k] = variable.ID
				break
			}
		}

		// Добавление последнего канала
		if i == len(sys.Itfs)-1 {
			ch := parsing.OuterChannel{
				Name:         mbName,
				Description:  "",
				Offset:       strconv.Itoa(startPos),
				Length:       strconv.Itoa(strDelta(itf.Offset, strconv.Itoa(startPos)) + 1),
				FunctionCode: fc2Text(itf.Func),
				CycleTime:    "1000",
				ChannelType:  "Timer",
			}
			mbXml.OuterChannels = append(mbXml.OuterChannels, ch)
		}

		prevFC = itf.Func
		variables = append(variables, variable)
	}

	// Генерация XML и сохранение данных
	xmlText, err := parsing.ToXML(mbXml)
	if err != nil {
		return err
	}
	sys.MbXml = string(xmlText)
	sys.Variables = variables
	sys.FuncBlocks = fbs

	return nil
}

// strDelta вычисляет разницу между двумя строковыми числами
func strDelta(cur, prev string) int {
	prevInt, _ := strconv.Atoi(prev)
	curInt, _ := strconv.Atoi(cur)
	return curInt - prevInt
}

// fc2Text преобразует код функции Modbus в текстовое представление
func fc2Text(fc string) string {
	switch fc {
	case "03":
		return "ReadHoldingRegisters"
	case "06":
		return "WriteHoldingRegisters"
	default:
		return "Error"
	}
}
