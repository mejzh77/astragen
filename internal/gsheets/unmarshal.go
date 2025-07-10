// Package gsheets
// Реализует парсер для Google-таблиц
package gsheets

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"
)

// Parser - основной интерфейс библиотеки
type Parser struct {
	headers    []any
	headerMap  map[string]int
	timeFormat string
}

// NewParser создает новый парсер для Google Sheets
func NewParser(headers []any) *Parser {
	p := &Parser{
		headers:   headers,
		headerMap: make(map[string]int),
	}

	// Создаем нормализованную карту заголовков
	for i, header := range headers {
		normalized := p.normalizeHeader(header.(string))
		p.headerMap[normalized] = i
	}

	return p
}

// parseStruct рекурсивно парсит структуру
func (p *Parser) parseStruct(row []any, val reflect.Value) error {
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		tag := field.Tag.Get("gsheets")
		// Обрабатываем вложенные структуры
		if field.Anonymous {
			if strings.Contains(tag, ",squash") || tag == "squash" {
				if err := p.parseStruct(row, fieldVal); err != nil {
					return err
				}
				continue
			}
		}
		// Пропускаем непэкспортируемые поля
		if !fieldVal.CanSet() {
			continue
		}

		if tag == "" || tag == "-" {
			continue
		}
		// Обработка опций тега (например, "required")
		tagParts := strings.Split(tag, ",")
		columnName := tagParts[0]
		options := make(map[string]bool)
		for _, opt := range tagParts[1:] {
			options[strings.TrimSpace(opt)] = true
		}

		// Получаем значение из строки
		value, exists := p.getValue(row, columnName)
		if !exists {
			if options["required"] {
				return fmt.Errorf("gsheets: required column '%s' not found", columnName)
			}
			continue
		}

		// Конвертируем значение
		if err := p.convertValue(value, fieldVal, options); err != nil {
			return fmt.Errorf("gsheets: error in column '%s': %w", columnName, err)
		}
	}

	return nil
}

// getValue возвращает значение из строки по имени колонки
func (p *Parser) getValue(row []any, columnName string) (string, bool) {
	normalized := p.normalizeHeader(columnName)
	idx, exists := p.headerMap[normalized]
	if !exists {
		return "", false
	}

	if idx >= len(row) {
		return "", false
	}

	return sanitizeString(strings.TrimSpace(row[idx].(string))), true
}

// Parse парсит строку в структуру
func (p *Parser) Parse(row []any, v any) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return &InvalidUnmarshalError{Type: reflect.TypeOf(v)}
	}

	// Дереференсируем указатель
	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return errors.New("gsheets: must pass a pointer to struct")
	}
	return p.parseStruct(row, val)
}

// Unmarshal парсит данные из Google Sheets в структуру
// rows - строки таблицы (первая строка - заголовки)
// v - указатель на слайс структур
func Unmarshal(rows [][]any, v any) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return &InvalidUnmarshalError{Type: reflect.TypeOf(v)}
	}

	// Получаем тип элементов слайса
	sliceVal := val.Elem()
	if sliceVal.Kind() != reflect.Slice {
		return errors.New("gsheets: target must be a slice")
	}

	elemType := sliceVal.Type().Elem()
	if elemType.Kind() != reflect.Struct {
		return errors.New("gsheets: slice elements must be structs")
	}

	// Обработка пустых данных
	if len(rows) == 0 {
		return nil
	}
	// Создаем карту: имя колонки -> индекс
	headers := rows[0]
	parser := NewParser(headers)
	parser.SetTimeFormat(time.RFC3339) // Устанавливаем формат времени
	// Парсим каждую строку
	for _, row := range rows[1:] {
		// Создаем новую структуру
		newElem := reflect.New(elemType).Elem()
		// Парсим строку
		if err := parser.Parse(row, newElem.Addr().Interface()); err != nil {
			panic(err)
		}
		// fmt.Println(newElem)
		//  Добавляем в слайс
		sliceVal.Set(reflect.Append(sliceVal, newElem))
	}

	return nil
}

// normalizeHeader приводит заголовок к нормализованному виду

func (p *Parser) normalizeHeader(header string) string {
	return strings.ToLower(strings.TrimSpace(header))
}

func sanitizeString(s string) string {

	if !utf8.ValidString(s) {

		log.Printf("WARNING: Invalid UTF-8 string found: %q", s)
		// Удаляем невалидные UTF-8 символы
		v := make([]rune, 0, len(s))
		for i, r := range s {
			if r == utf8.RuneError {
				_, size := utf8.DecodeRuneInString(s[i:])
				if size == 1 {
					continue // Пропускаем невалидный символ
				}
			}
			v = append(v, r)
		}
		s = string(v)
	}
	return parseCellValue(s)
}
func parseCellValue(value interface{}) string {
	str := fmt.Sprintf("%v", value)
	// Удаляем непечатные символы
	str = strings.Map(func(r rune) rune {
		if r >= 32 && r != 127 {
			return r
		}
		return -1
	}, str)
	return str
}
