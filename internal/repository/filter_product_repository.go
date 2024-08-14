package repository

import (
	"context"
	"weblineBackend/internal/model"
)

type FilterProductRepository interface {
	GetTotalProductsByFilters(ctx context.Context, filterValues *model.AllProductFilterValues) (int64, error)
	GetProductsByFilters(ctx context.Context, filterValues *model.AllProductFilterValues) ([]*model.Product, error)
	GetProductAttributes(ctx context.Context) (*model.FilterOptions, error)
}
