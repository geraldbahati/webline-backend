package repository

import (
	"context"
	"github.com/google/uuid"
	"weblineBackend/internal/model"
)

type FilterCategoryProductRepository interface {
	GetTotalCategoryProductsByFilters(ctx context.Context, filterValues *model.CategoryProductFilterValues) (int64, error)
	GetCategoryProductsByFilters(ctx context.Context, filterValues *model.CategoryProductFilterValues) ([]*model.Product, error)
	GetProductAttributesAndCountByCategoryID(ctx context.Context, categoryID uuid.UUID) (*model.FilterOptions, error)
}
