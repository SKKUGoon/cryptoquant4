package utils

import (
	"fmt"
	"math"

	"github.com/shopspring/decimal"
)

func SafeDecimalFromFloat(f float64) (decimal.Decimal, error) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("decimal.NewFromFloat panic: %v. value was %f", r, f)
		}
	}()

	if math.IsNaN(f) || math.IsInf(f, 0) {
		return decimal.Zero, fmt.Errorf("invalid float value: %f", f)
	}

	return decimal.NewFromFloat(f), err
}
