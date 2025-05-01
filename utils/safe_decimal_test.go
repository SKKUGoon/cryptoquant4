package utils_test

import (
	"math"
	"testing"

	"cryptoquant.com/m/utils"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestSafeDecimalFromFloat(t *testing.T) {
	tests := []struct {
		name      string
		input     float64
		wantValue decimal.Decimal
		wantErr   bool
	}{
		{
			name:      "Valid float",
			input:     123.456,
			wantValue: decimal.NewFromFloat(123.456),
			wantErr:   false,
		},
		{
			name:      "Zero float",
			input:     0.0,
			wantValue: decimal.Zero,
			wantErr:   false,
		},
		{
			name:      "NaN",
			input:     math.NaN(),
			wantValue: decimal.Zero,
			wantErr:   true,
		},
		{
			name:      "Positive Infinity",
			input:     math.Inf(1),
			wantValue: decimal.Zero,
			wantErr:   true,
		},
		{
			name:      "Negative Infinity",
			input:     math.Inf(-1),
			wantValue: decimal.Zero,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.SafeDecimalFromFloat(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, got.Equal(tt.wantValue), "Expected zero decimal on error")
			} else {
				assert.NoError(t, err)
				assert.True(t, got.Equal(tt.wantValue), "Expected %v, got %v", tt.wantValue, got)
			}
		})
	}
}

func TestSafeDecimalFromFloatTruncate(t *testing.T) {
	tests := []struct {
		name      string
		input     float64
		precision int
		wantValue decimal.Decimal
		wantErr   bool
	}{
		{
			name:      "Valid float, precision 2",
			input:     123.456789,
			precision: 2,
			wantValue: decimal.NewFromFloat(123.45),
			wantErr:   false,
		},
		{
			name:      "Valid float, precision 0",
			input:     123.456,
			precision: 0,
			wantValue: decimal.NewFromFloat(123),
			wantErr:   false,
		},
		{
			name:      "Valid float, high precision",
			input:     123.456,
			precision: 5,
			wantValue: decimal.NewFromFloat(123.456),
			wantErr:   false,
		},
		{
			name:      "Zero float",
			input:     0.0,
			precision: 3,
			wantValue: decimal.Zero,
			wantErr:   false,
		},
		{
			name:      "NaN",
			input:     math.NaN(),
			precision: 2,
			wantValue: decimal.Zero,
			wantErr:   true,
		},
		{
			name:      "Positive Infinity",
			input:     math.Inf(1),
			precision: 2,
			wantValue: decimal.Zero,
			wantErr:   true,
		},
		{
			name:      "Negative Infinity",
			input:     math.Inf(-1),
			precision: 2,
			wantValue: decimal.Zero,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.SafeDecimalFromFloatTruncate(tt.input, tt.precision)

			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, got.Equal(tt.wantValue), "Expected zero decimal on error")
			} else {
				assert.NoError(t, err)
				// Compare string representations because float conversion can have small inaccuracies
				assert.Equal(t, tt.wantValue.String(), got.String(), "Expected %v, got %v", tt.wantValue.String(), got.String())
			}
		})
	}
}
