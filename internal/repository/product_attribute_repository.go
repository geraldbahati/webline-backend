package repository

import (
	"context"
	"github.com/google/uuid"
	"weblineBackend/internal/model"
)

type ProductAttributeRepository interface {
	CreateProductAttribute(ctx context.Context, name, attributeType string) (*uuid.UUID, error)
	CreateProductAttributeValue(ctx context.Context, attributeID uuid.UUID, categoryID uuid.NullUUID, value, hexValue string) (*uuid.UUID, error)
	CreateProductToAttributeValue(ctx context.Context, productID uuid.UUID, attributeValueID uuid.UUID) error
	GetProductAttributesWithValues(ctx context.Context, categoryID uuid.UUID) (map[string][]model.Attribute, error)
}
