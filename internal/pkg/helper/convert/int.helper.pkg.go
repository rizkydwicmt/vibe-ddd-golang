package convert

import (
	"reflect"
	"strconv"

	"github.com/samber/lo"
)

func GetMapIntValue(header map[string]interface{}, key string) *int {
	v := 0
	value, exists := header[key]
	if !exists || value == nil {
		return &v
	}

	aValue := reflect.ValueOf(value)
	switch aValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v = int(aValue.Int())
	case reflect.Float32, reflect.Float64:
		v = int(aValue.Float())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v = int(aValue.Uint())
	case reflect.Bool:
		if aValue.Bool() {
			v = 1
		}
	case reflect.String:
		parsed, err := strconv.Atoi(aValue.String())
		return lo.Ternary(err == nil, &parsed, &v)
	default:
		return &v
	}

	return &v
}

func GetMapInt64Value(header map[string]interface{}, key string) *int64 {
	v := int64(0)
	value, exists := header[key]
	if !exists || value == nil {
		return &v
	}

	aValue := reflect.ValueOf(value)
	switch aValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v = aValue.Int()
	case reflect.Float32, reflect.Float64:
		v = int64(aValue.Float())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v = int64(aValue.Uint())
	case reflect.Bool:
		if aValue.Bool() {
			v = 1
		}
	case reflect.String:
		parsed, err := strconv.ParseInt(aValue.String(), 10, 64)
		return lo.Ternary(err == nil, &parsed, &v)
	default:
		return &v
	}

	return &v
}

func GetMapInt64NilValue(header map[string]interface{}, key string) *int64 {
	v := int64(0)
	value, exists := header[key]
	if !exists || value == nil {
		return nil
	}

	aValue := reflect.ValueOf(value)
	switch aValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v = aValue.Int()
	case reflect.Float32, reflect.Float64:
		v = int64(aValue.Float())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v = int64(aValue.Uint())
	case reflect.Bool:
		if aValue.Bool() {
			v = 1
		}
	case reflect.String:
		parsed, err := strconv.ParseInt(aValue.String(), 10, 64)
		return lo.Ternary(err == nil, &parsed, nil)
	default:
		return nil
	}

	return &v
}
