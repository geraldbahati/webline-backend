package repository

import (
	"context"
	"weblineBackend/internal/model"
)

// SearchRepository defines methods for product search functionality based on search.sql.
type SearchRepository interface {
	// V2SearchProducts performs a full-text search on products.
	// It searches based on the provided searchTerm and limits the number of returned results.
	V2SearchProducts(ctx context.Context, searchTerm string, limit int) ([]model.SearchProductResult, error)

	// AutocompleteSuggestions returns distinct auto-complete suggestions
	// based on product and category names.
	AutocompleteSuggestions(ctx context.Context, searchTerm string, limit int) ([]string, error)

	// V2SearchProductsPaginated performs a full-text search on products with pagination support.
	// It returns a PaginationResult containing the paginated search results.
	V2SearchProductsPaginated(ctx context.Context, searchTerm string, page int, pageSize int) (model.PaginationResult[[]model.SearchProductResult], error)
}
