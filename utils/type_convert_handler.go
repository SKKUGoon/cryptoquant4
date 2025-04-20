package utils

import "strconv"

// Example converters
func StringToFloat64(v string) float64 {
	f, _ := strconv.ParseFloat(v, 64)
	return f
}

func StringToString(v string) string {
	return v
}

func ConvertToAnySlice[T any](slice []T) []any {
	result := make([]any, len(slice))
	for i, v := range slice {
		result[i] = v
	}
	return result
}
