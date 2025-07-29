package gsheets

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// Marshal преобразует структуру или слайс структур в формат, пригодный для записи в Google Sheets
func Marshal(data interface{}) ([][]interface{}, error) {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Struct:
		return marshalStruct(val)
	case reflect.Slice:
		return marshalSlice(val)
	default:
		return nil, fmt.Errorf("gsheets: expected struct or slice, got %T", data)
	}
}

// marshalStruct преобразует структуру в строку для Google Sheets
func marshalStruct(val reflect.Value) ([][]interface{}, error) {
	if val.Kind() != reflect.Struct {
		return nil, errors.New("gsheets: expected struct")
	}

	headers, values, err := extractFields(val)
	if err != nil {
		return nil, err
	}

	return [][]interface{}{headers, values}, nil
}

// marshalSlice преобразует слайс структур в данные для Google Sheets
func marshalSlice(val reflect.Value) ([][]interface{}, error) {
	if val.Kind() != reflect.Slice {
		return nil, errors.New("gsheets: expected slice")
	}

	if val.Len() == 0 {
		return nil, nil
	}

	// Проверяем, что элементы - структуры
	elemType := val.Type().Elem()
	if elemType.Kind() != reflect.Struct {
		return nil, errors.New("gsheets: slice elements must be structs")
	}

	var result [][]interface{}

	// Добавляем заголовки из первой структуры
	if val.Len() > 0 {
		headers, _, err := extractFields(val.Index(0))
		if err != nil {
			return nil, err
		}
		result = append(result, headers)
	}

	// Добавляем значения для каждой структуры
	for i := 0; i < val.Len(); i++ {
		_, values, err := extractFields(val.Index(i))
		if err != nil {
			return nil, err
		}
		result = append(result, values)
	}

	return result, nil
}

// extractFields извлекает поля из структуры согласно тегам gsheets
func extractFields(val reflect.Value) ([]interface{}, []interface{}, error) {
	var headers []interface{}
	var values []interface{}

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		tag := field.Tag.Get("gsheets")
		if tag == "-" {
			continue
		}

		// Обрабатываем вложенные структуры
		if field.Type.Kind() == reflect.Struct && field.Type != reflect.TypeOf(time.Time{}) {
			if strings.Contains(tag, ",squash") || tag == "squash" {
				nestedHeaders, nestedValues, err := extractFields(fieldVal)
				if err != nil {
					return nil, nil, err
				}
				headers = append(headers, nestedHeaders...)
				values = append(values, nestedValues...)
				continue
			}
		}

		// Пропускаем непэкспортируемые поля
		if !fieldVal.CanInterface() {
			continue
		}

		if tag == "" {
			continue
		}

		// Извлекаем имя колонки из тега
		columnName := strings.Split(tag, ",")[0]
		headers = append(headers, columnName)

		// Преобразуем значение в подходящий формат
		value, err := convertToSheetValue(fieldVal)
		if err != nil {
			return nil, nil, fmt.Errorf("field %s: %w", field.Name, err)
		}
		values = append(values, value)
	}

	return headers, values, nil
}

// convertToSheetValue преобразует значение поля в формат для Google Sheets
func convertToSheetValue(fieldVal reflect.Value) (interface{}, error) {
	switch fieldVal.Kind() {
	case reflect.String:
		return fieldVal.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fieldVal.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fieldVal.Uint(), nil
	case reflect.Float32, reflect.Float64:
		return fieldVal.Float(), nil
	case reflect.Bool:
		return fieldVal.Bool(), nil
	case reflect.Struct:
		if t, ok := fieldVal.Interface().(time.Time); ok {
			return t.Format(time.RFC3339), nil
		}
		return nil, fmt.Errorf("unsupported struct type: %T", fieldVal.Interface())
	case reflect.Ptr:
		if fieldVal.IsNil() {
			return nil, nil
		}
		return convertToSheetValue(fieldVal.Elem())
	default:
		return nil, fmt.Errorf("unsupported type: %s", fieldVal.Kind())
	}
}
