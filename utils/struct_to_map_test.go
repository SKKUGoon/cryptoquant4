package utils_test

import (
	"testing"

	"cryptoquant.com/m/utils"
	"github.com/shopspring/decimal"
)

type TestStruct struct {
	StringField    string  `json:"string_field"`
	IntField       int     `json:"int_field"`
	FloatField     float64 `json:"float_field"`
	BoolField      bool    `json:"bool_field"`
	OmitEmptyField string  `json:"omit_empty_field,omitempty"`
	NoTagField     string
	DecimalField   decimal.Decimal `json:"decimal_field"`
}

func TestStructToParamsMap(t *testing.T) {
	tests := []struct {
		name     string
		input    TestStruct
		expected map[string]string
	}{
		{
			name: "All fields populated",
			input: TestStruct{
				StringField:    "test",
				IntField:       42,
				FloatField:     3.14,
				BoolField:      true,
				OmitEmptyField: "not empty",
				NoTagField:     "no tag",
				DecimalField:   decimal.NewFromFloat(123.45),
			},
			expected: map[string]string{
				"string_field":     "test",
				"int_field":        "42",
				"float_field":      "3.14",
				"bool_field":       "true",
				"omit_empty_field": "not empty",
				"decimal_field":    "123.45",
			},
		},
		{
			name: "Empty fields with omitempty",
			input: TestStruct{
				StringField:  "test",
				IntField:     42,
				FloatField:   3.14,
				BoolField:    true,
				NoTagField:   "no tag",
				DecimalField: decimal.NewFromFloat(123.45),
			},
			expected: map[string]string{
				"string_field":  "test",
				"int_field":     "42",
				"float_field":   "3.14",
				"bool_field":    "true",
				"decimal_field": "123.45",
			},
		},
		{
			name: "Zero values",
			input: TestStruct{
				StringField:    "",
				IntField:       0,
				FloatField:     0,
				BoolField:      false,
				OmitEmptyField: "",
				NoTagField:     "",
				DecimalField:   decimal.NewFromFloat(0),
			},
			expected: map[string]string{
				"string_field":  "",
				"int_field":     "0",
				"float_field":   "0",
				"bool_field":    "false",
				"decimal_field": "0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.StructToParamsMap(&tt.input)

			// Check if all expected keys are present with correct values
			for key, expectedValue := range tt.expected {
				if actualValue, exists := result[key]; !exists {
					t.Errorf("Expected key %s not found in result", key)
				} else if actualValue != expectedValue {
					t.Errorf("For key %s, expected value %s, got %s", key, expectedValue, actualValue)
				}
			}

			// Check if there are no unexpected keys in the result
			for key := range result {
				if _, exists := tt.expected[key]; !exists {
					t.Errorf("Unexpected key %s found in result", key)
				}
			}
		})
	}
}

func TestStructToParamsMapEdgeCases(t *testing.T) {
	t.Run("nil pointer", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for nil pointer but got none")
			}
		}()
		var nilStruct *TestStruct
		_ = utils.StructToParamsMap(nilStruct)
	})

	t.Run("non-struct type", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for non-struct type but got none")
			}
		}()
		var nonStruct int
		_ = utils.StructToParamsMap(&nonStruct)
	})
}
