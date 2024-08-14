package utils

import (
	"math"
)

// RoundOffPrice rounds off the price according to the specified rules.
func RoundOffPrice(price float64) float64 {
	if price == 0 {
		return 0
	}

	// Determine the number of digits in the integer part of the price
	magnitude := math.Floor(math.Log10(price))

	// Calculate the rounding factor (e.g., 10, 100, 1000, etc.)
	roundingFactor := math.Pow(10, magnitude-1) // Adjusting to round at the next significant place

	// Round the price up to the nearest rounding factor
	roundedPrice := math.Ceil(price/roundingFactor) * roundingFactor

	return roundedPrice
}
