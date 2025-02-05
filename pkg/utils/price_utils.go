package utils

import (
	"math"
	"strconv"
)

// RoundPrice rounds the price to the nearest 10 if it's less than 100,
// or to the nearest 100 if it's 100 or more.
func RoundPrice(price float64) float64 {
	if price < 100 {
		return math.Round(price/10) * 10
	}
	return math.Round(price/100) * 100
}

func RoundPriceString(price string) string {
	priceFloat, err := strconv.ParseFloat(price, 64)
	if err != nil {
		return price
	}
	if priceFloat < 100 {
		return strconv.FormatFloat(math.Round(priceFloat/10)*10, 'f', -1, 64)
	}
	return strconv.FormatFloat(math.Round(priceFloat/100)*100, 'f', -1, 64)
}

// ConvertUSDToKES converts a USD price (as a string) to Kenyan Shillings
// using the provided exchangeRate. It returns the converted price rounded using RoundPrice.
func ConvertUSDToKES(usdPrice string, exchangeRate float64) (float64, error) {
	price, err := strconv.ParseFloat(usdPrice, 64)
	if err != nil {
		return 0, err
	}
	converted := price * exchangeRate
	return RoundPrice(converted), nil
}
