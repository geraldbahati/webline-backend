package repository

import (
	"context"
	"github.com/google/uuid"
)

type ProductAttributeRepository interface {
	CreateProductAttribute(ctx context.Context, name, attributeType string) (*uuid.UUID, error)
	CreateProductAttributeValue(ctx context.Context, attributeID uuid.UUID, categoryID uuid.NullUUID, value, hexValue string) (*uuid.UUID, error)
	CreateProductToAttributeValue(ctx context.Context, productID uuid.UUID, attributeValueID uuid.UUID) error
}
