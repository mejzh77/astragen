package gsheets

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Unmarshal парсит данные из Google Sheets в структуру
// rows - строки таблицы (первая строка - заголовки)
// v - указатель на слайс структур
func Unmarshal(rows [][]interface{}, v interface{}) error {
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
	headerMap := make(map[string]int)
	for i, header := range headers {
		normalized := normalizeHeader(header.(string))
		headerMap[normalized] = i
	}

	// Парсим каждую строку
	for _, row := range rows[1:] {
		// Создаем новую структуру
		newElem := reflect.New(elemType).Elem()

		// Заполняем поля структуры
		for i := 0; i < elemType.NumField(); i++ {
			field := elemType.Field(i)
			tag := field.Tag.Get("gsheets")
			if tag == "" || tag == "-" {
				continue
			}

			// Ищем индекс колонки
			normalizedTag := normalizeHeader(tag)
			colIndex, exists := headerMap[normalizedTag]
			if !exists {
				return &UnknownFieldError{Field: tag}
			}

			// Получаем значение из строки
			var value string
			if colIndex < len(row) {
				value = row[colIndex].(string)
			}

			// Конвертируем и устанавливаем значение
			if err := convertValue(value, newElem.Field(i)); err != nil {
				return fmt.Errorf("gsheets: error in column '%s': %w", tag, err)
			}
		}

		// Добавляем в слайс
		sliceVal.Set(reflect.Append(sliceVal, newElem))
	}

	return nil
}

// normalizeHeader приводит заголовок к нормализованному виду
func normalizeHeader(header string) string {
	return strings.ToLower(strings.TrimSpace(header))
}
