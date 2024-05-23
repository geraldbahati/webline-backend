package model

import "time"

type ProductImage struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	S3URL     string    `json:"s3_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
