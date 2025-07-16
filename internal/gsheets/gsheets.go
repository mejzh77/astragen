package gsheets

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	"google.golang.org/api/sheets/v4"
)

type Service struct {
	client *sheets.Service
}

func NewService(ctx context.Context, credentialsJSON []byte) (*Service, error) {
	client, err := google.JWTConfigFromJSON(
		credentialsJSON,
		sheets.SpreadsheetsReadonlyScope,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT config: %w", err)
	}

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client.Client(ctx)))
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets service: %w", err)
	}

	return &Service{client: srv}, nil
}

func (s *Service) ReadSheet(spreadsheetID, readRange string) ([][]interface{}, error) {
	resp, err := s.client.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to read sheet: %w", err)
	}
	return resp.Values, nil
}

// GetRange формирует диапазон для чтения данных из Google Sheets
// sheetName - имя листа
// structTemplate - структура для парсинга данных
// withHeader - включать ли строку заголовков
func GetRange(sheetName string, structTemplate interface{}, withHeader bool) (string, error) {
	// Проверяем, что передана структура
	val := reflect.ValueOf(structTemplate)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return "", fmt.Errorf("expected struct, got %T", structTemplate)
	}

	// Определяем количество столбцов
	numColumns := calculateColumns(val.Type())

	// Формируем диапазон
	startCell := "A1"
	if !withHeader {
		startCell = "A2"
	}

	// Если количество столбцов больше 26 (Z), используем нотацию типа AA, AB и т.д.
	endColumn := columnToLetter(numColumns)
	// Для больших таблиц можно ограничить количество строк
	// Например, 1000 строк данных + 1 строка заголовка
	return fmt.Sprintf("%s!%s:%s", sheetName, startCell, endColumn), nil
}
func getAllHeaders(service *sheets.Service, spreadsheetID, sheetName string) ([]string, error) {
	// Читаем только первую строку полностью
	readRange := fmt.Sprintf("%s!1:1", sheetName)
	resp, err := service.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve headers: %v", err)
	}

	if len(resp.Values) == 0 {
		return nil, errors.New("no headers found")
	}

	var headers []string
	for _, cell := range resp.Values[0] {
		headers = append(headers, fmt.Sprintf("%v", cell))
	}

	return headers, nil
}

// calculateColumns рекурсивно вычисляет количество столбцов в структуре
func calculateColumns(typ reflect.Type) int {
	var count int

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// Пропускаем непэкспортируемые поля
		if field.PkgPath != "" {
			continue
		}

		// Проверяем тег gsheets
		tag := field.Tag.Get("gsheets")
		if tag == "-" {
			continue
		}

		// Обрабатываем вложенные структуры
		if field.Type.Kind() == reflect.Struct && field.Type != reflect.TypeOf(time.Time{}) {
			count += calculateColumns(field.Type)
			continue
		}

		// Если тег не пустой и не "-", считаем как отдельную колонку
		if tag != "" {
			count++
		}
	}

	return count
}

// В internal/gsheets/service.go
func (s *Service) Load(spreadsheetID string, sheetName string, dest interface{}) error {
	//val := reflect.ValueOf(dest)
	//if val.Kind() != reflect.Ptr || val.IsNil() {
	//return &InvalidUnmarshalError{Type: reflect.TypeOf(dest)}
	//}

	//// Получаем тип элементов слайса
	//sliceVal := val.Elem()
	//if sliceVal.Kind() != reflect.Slice {
	//return errors.New("gsheets: target must be a slice")
	//}

	//elemType := sliceVal.Type().Elem()
	//if elemType.Kind() != reflect.Struct {
	//return errors.New("gsheets: slice elements must be structs")
	//}
	//// Создаем экземпляр элемента для GetRange
	//elem := reflect.New(elemType).Elem().Interface()
	// 1. Определяем диапазон для чтения
	// 1. Получаем ВСЕ заголовки из таблицы
	allHeaders, err := getAllHeaders(s.client, spreadsheetID, sheetName)
	if err != nil {
		return fmt.Errorf("failed to get headers: %w", err)
	}

	// 2. Создаем карту заголовков
	headerMap := make(map[string]int)
	for i, header := range allHeaders {
		headerMap[strings.TrimSpace(header)] = i
	}

	// 3. Определяем нужные колонки из структуры
	neededColumns := getNeededColumns(dest)

	// 4. Формируем диапазон чтения (все строки, только нужные колонки)
	readRange := buildRange(sheetName, neededColumns, headerMap)
	// 2. Читаем данные из таблицы
	rows, err := s.ReadSheet(spreadsheetID, readRange)
	if err != nil {
		return fmt.Errorf("failed to read sheet: %w", err)
	}

	// 3. Парсим данные в целевую структуру
	if err := Unmarshal(rows, dest); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}
func getNeededColumns(dest interface{}) map[string]bool {
	needed := make(map[string]bool)
	val := reflect.ValueOf(dest).Elem()
	typ := val.Type().Elem() // Тип элемента среза

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("gsheets")
		if tag == "" || tag == "-" {
			continue
		}
		needed[strings.Split(tag, ",")[0]] = true
	}
	return needed
}

func buildRange(sheetName string, neededColumns map[string]bool, headerMap map[string]int) string {
	var columns []string
	for colName := range neededColumns {
		if idx, exists := headerMap[colName]; exists {
			colLetter := toColumnLetter(idx + 1) // +1 т.к. индексы с 0
			columns = append(columns, colLetter)
		}
	}
	sort.Strings(columns) // Сортируем для правильного порядка

	if len(columns) == 0 {
		return fmt.Sprintf("%s!A:Z", sheetName) // Дефолтный диапазон
	}

	return fmt.Sprintf("%s!%s:%s", sheetName, columns[0], columns[len(columns)-1])
}

func toColumnLetter(colNum int) string {
	letter := ""
	for colNum > 0 {
		colNum--
		letter = string(rune('A'+(colNum%26))) + letter
		colNum /= 26
	}
	return letter
}

// columnToLetter преобразует номер колонки в буквенное обозначение (1 -> A, 26 -> Z, 27 -> AA)
func columnToLetter(col int) string {
	letter := ""
	for col > 0 {
		col--
		letter = string(rune('A'+(col%26))) + letter
		col = col / 26
	}
	return letter
}
