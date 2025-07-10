package gsheets

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"gorm.io/gorm"

	"github.com/mejzh77/astragen/configs/config"
	"github.com/mejzh77/astragen/internal/repository"
	"github.com/mejzh77/astragen/pkg/models"
	"google.golang.org/api/sheets/v4"
)

type GoogleSheetsService struct {
	service    *sheets.Service
	signalRepo *repository.SignalRepository
}

func NewGoogleSheetsService(client *sheets.Service, db *gorm.DB) *GoogleSheetsService {
	return &GoogleSheetsService{
		service:    client,
		signalRepo: repository.NewSignalRepository(db),
	}
}

func NewService(ctx context.Context, credentialsJSON []byte) (*sheets.Service, error) {
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

	return srv, nil
}

func (s *GoogleSheetsService) ReadSheet(spreadsheetID, readRange string) ([][]interface{}, error) {
	resp, err := s.service.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to read sheet: %w", err)
	}
	return resp.Values, nil
}

func (s *GoogleSheetsService) RunSync(ctx context.Context) error {
	var allSignals []models.Signal
	for _, sheetCfg := range config.AppConfig.Sheets {
		// 1. Чтение данных из листа
		readRange, err := GetRange(sheetCfg.SheetName, sheetCfg.Model, true)
		if err != nil {
			return fmt.Errorf("failed to GetRange for sheet %s: %w", sheetCfg.SheetName, err)
		}

		rows, err := s.ReadSheet(config.AppConfig.SpreadsheetID, readRange)
		if err != nil {
			return fmt.Errorf("failed to read sheet %s: %w", sheetCfg.SheetName, err)
		}

		// 2. Парсинг в соответствующую модель
		modelSlice := reflect.New(reflect.SliceOf(reflect.TypeOf(sheetCfg.Model).Elem()))
		if err := Unmarshal(rows, modelSlice.Interface()); err != nil {
			return fmt.Errorf("failed to unmarshal sheet %s: %w", sheetCfg.SheetName, err)
		}

		// 3. Преобразование в Signal
		for i := 0; i < modelSlice.Elem().Len(); i++ {
			item := modelSlice.Elem().Index(i).Addr().Interface()
			var signal models.Signal

			switch v := item.(type) {
			case *models.DI:
				signal.FromDI(*v)
			case *models.AI:
				signal.FromAI(*v)
			case *models.DO:
				signal.FromDO(*v)
			case *models.AO:
				signal.FromAO(*v)
			}

			signal.SignalType = sheetCfg.SheetName
			allSignals = append(allSignals, signal)
		}
	}

	// 4. Сохранение в БД
	if err := s.signalRepo.SaveSignals(allSignals); err != nil {
		return fmt.Errorf("failed to save signals: %w", err)
	}

	return nil
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
