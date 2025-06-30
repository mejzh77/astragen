package gsheets

import (
	"fmt"
	"reflect"
)

type UnmarshalTypeError struct {
	Value string
	Type  reflect.Type
}

func (e *UnmarshalTypeError) Error() string {
	return fmt.Sprintf("gsheets: cannot unmarshal %q into Go value of type %s", e.Value, e.Type)
}

type UnknownFieldError struct {
	Field string
}

func (e *UnknownFieldError) Error() string {
	return fmt.Sprintf("gsheets: unknown field '%s' in sheet", e.Field)
}

type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "gsheets: Unmarshal(nil)"
	}
	return "gsheets: Unmarshal(non-pointer " + e.Type.String() + ")"
}
// newUnmarshalTypeError создает ошибку несоответствия типа
func newUnmarshalTypeError(value string, typ reflect.Type) error {
	return &UnmarshalTypeError{
		Value: value,
		Type:  typ,
	}
}
