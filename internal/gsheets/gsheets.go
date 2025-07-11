package gsheets

import (
	"context"
	"fmt"
	"reflect"
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
