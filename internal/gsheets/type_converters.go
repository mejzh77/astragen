package gsheets

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var timeFormats = []string{
	time.RFC3339,
	"2006-01-02",
	"2006-01-02 15:04:05",
	"02.01.2006",
	"02.01.2006 15:04:05",
}

func convertValue(value string, field reflect.Value) error {
	if value == "" {
		return nil // Пустые значения не меняют состояние поля
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return &UnmarshalTypeError{Value: value, Type: field.Type()}
		}
		field.SetInt(intVal)
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return &UnmarshalTypeError{Value: value, Type: field.Type()}
		}
		field.SetUint(uintVal)
		return nil

	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return &UnmarshalTypeError{Value: value, Type: field.Type()}
		}
		field.SetFloat(floatVal)
		return nil

	case reflect.Bool:
		boolVal, err := parseBool(value)
		if err != nil {
			return &UnmarshalTypeError{Value: value, Type: field.Type()}
		}
		field.SetBool(boolVal)
		return nil

	case reflect.Struct:
		if field.Type() == reflect.TypeOf(time.Time{}) {
			t, err := parseTime(value)
			if err != nil {
				return &UnmarshalTypeError{Value: value, Type: field.Type()}
			}
			field.Set(reflect.ValueOf(t))
			return nil
		}
	}

	return fmt.Errorf("gsheets: unsupported type %s", field.Type())
}

func parseBool(value string) (bool, error) {
	switch strings.ToLower(value) {
	case "true", "1", "yes", "y", "on":
		return true, nil
	case "false", "0", "no", "n", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value")
	}
}

func parseTime(value string) (time.Time, error) {
	for _, format := range timeFormats {
		t, err := time.Parse(format, value)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized time format")
}
