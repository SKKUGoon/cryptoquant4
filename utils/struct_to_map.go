package utils

import (
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

func StructToParamsMap(s any) map[string]string {
	result := make(map[string]string)

	// Get addressable struct.
	// To get `Addr` safely, we need the original object, not a pointer or a copy
	v := reflect.ValueOf(s).Elem()

	// Gets type information of the struct.
	// Extract Struct fields and tags (for json)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)       // Value of the field
		structField := t.Field(i) // Struct field definition

		jsonTag := structField.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			// No tag
			continue
		}
		jsonKey := parseJSONKey(jsonTag)

		if hasOmitEmpty(jsonTag) && field.IsZero() {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			result[jsonKey] = field.String()
		case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
			result[jsonKey] = strconv.FormatInt(field.Int(), 10)
		case reflect.Float32, reflect.Float64:
			result[jsonKey] = strconv.FormatFloat(field.Float(), 'f', -1, 64)
		case reflect.Bool:
			result[jsonKey] = strconv.FormatBool(field.Bool())
		case reflect.Struct:
			// Handle decimal.Decimal
			if field.Type().String() == "decimal.Decimal" {
				ptr := field.Addr().Interface().(*decimal.Decimal)
				result[jsonKey] = ptr.String()
			}
		}
	}
	return result
}

// parseJSONKey extracts the key name from a JSON tag string by returning everything
// before the first comma if one exists, otherwise returns the full tag
func parseJSONKey(tag string) string {
	if idx := indexByte(tag, ','); idx != -1 {
		return tag[:idx]
	}
	return tag
}

// indexByte returns the index of the first occurrence of byte b in string s,
// or -1 if b is not present in s
func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

func hasOmitEmpty(tag string) bool {
	parts := strings.Split(tag, ",")
	return slices.Contains(parts, "omitempty")
}
