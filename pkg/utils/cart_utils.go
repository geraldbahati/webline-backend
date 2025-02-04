package utils

import (
	"fmt"
	"weblineBackend/internal/model"
)

// TransformCartItems iterates over each CartItem and updates the ImageURL to the full S3 URL
// as well as converting the Price (assumed stored as a USD string) to Kenyan Shillings (KES)
// using the provided exchangeRate. The price is formatted as a string with two decimals.
func TransformCartItems(cartItems []*model.CartItem, bucketName, region string, exchangeRate float64) {
	for _, item := range cartItems {
		// Update the image URL if it exists.
		if item.ImageURL != "" {
			item.ImageURL = ConstructS3URL(bucketName, region, item.ImageURL)
		}
		// Convert the Price from USD to KES if an input exists.
		if item.Price != "" {
			if converted, err := ConvertUSDToKES(item.Price, exchangeRate); err == nil {
				// Format the converted price as a string with two decimals.
				item.Price = fmt.Sprintf("%.2f", converted)
			}
		}
	}
}
