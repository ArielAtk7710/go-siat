package utils

import (
	"github.com/shopspring/decimal"
)

// Round rounds a float64 value to the specified number of decimals.
func Round(v float64, decimals int) float64 {
	return decimal.NewFromFloat(v).Round(int32(decimals)).InexactFloat64()
}
