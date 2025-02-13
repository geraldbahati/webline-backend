package sqlc

import (
	"context"
	"fmt"

	"weblineBackend/internal/database"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"

	"go.uber.org/zap"
)

// searchRepositoryImpl implements repository.SearchRepository using SQLC-generated queries.
type searchRepositoryImpl struct {
	*database.Queries
	logger *zap.Logger
}

// NewSearchRepositoryImpl returns a new instance of SearchRepository backed by SQLC.
func NewSearchRepositoryImpl(db *database.Queries, logger *zap.Logger) repository.SearchRepository {
	return &searchRepositoryImpl{
		Queries: db,
		logger:  logger,
	}
}

// V2SearchProducts executes the V2SearchProducts query from search.sql and converts the result rows
// into model.SearchProductResult.
func (r *searchRepositoryImpl) V2SearchProducts(ctx context.Context, searchTerm string, limit int) ([]model.SearchProductResult, error) {
	// Convert limit to int32 as expected by the generated code.
	rows, err := r.Queries.V2SearchProducts(ctx, database.V2SearchProductsParams{
		RegexpSplitToTable: searchTerm,
		Limit:              int32(limit),
	})
	if err != nil {
		r.logger.Error("failed to search products", zap.Error(err))
		return nil, fmt.Errorf("search products: %w", err)
	}

	results := make([]model.SearchProductResult, 0, len(rows))
	for _, row := range rows {
		results = append(results, model.SearchProductResult{
			ID:         row.ID,
			Name:       row.Name,
			PriceInKES: float64(row.PriceInKes),
			Slug:       row.Slug,
			Rank:       float64(row.Rank),
		})
	}
	return results, nil
}

// AutocompleteSuggestions calls the AutocompleteSuggestions query from search.sql.
func (r *searchRepositoryImpl) AutocompleteSuggestions(ctx context.Context, searchTerm string, limit int) ([]string, error) {
	suggestions, err := r.Queries.AutocompleteSuggestions(ctx, database.AutocompleteSuggestionsParams{
		Lower: searchTerm,
		Limit: int32(limit),
	})
	if err != nil {
		r.logger.Error("failed to fetch autocomplete suggestions", zap.Error(err))
		return nil, fmt.Errorf("autocomplete suggestions: %w", err)
	}
	return suggestions, nil
}

// V2SearchProductsPaginated executes the paginated search query.
func (r *searchRepositoryImpl) V2SearchProductsPaginated(ctx context.Context, searchTerm string, page int, pageSize int) (model.PaginationResult[[]model.SearchProductResult], error) {
	// Calculate offset: page numbers are 1-based.
	offset := int32((page - 1) * pageSize)
	limit := int32(pageSize)

	// Call the SQLC query that supports pagination.
	rows, err := r.Queries.V2SearchProductsPaginated(ctx, database.V2SearchProductsPaginatedParams{
		RegexpSplitToTable: searchTerm,
		Limit:              limit,
		Offset:             offset,
	})
	if err != nil {
		r.logger.Error("failed to fetch paginated search products", zap.Error(err))
		return model.PaginationResult[[]model.SearchProductResult]{}, fmt.Errorf("paginated search products: %w", err)
	}

	results := make([]model.SearchProductResult, 0, len(rows))
	for _, row := range rows {
		results = append(results, model.SearchProductResult{
			ID:         row.ID,
			Name:       row.Name,
			PriceInKES: float64(row.PriceInKes),
			Slug:       row.Slug,
			Rank:       float64(row.Rank),
		})
	}

	// Query to get the total count for the search term.
	count, err := r.Queries.V2SearchProductsCount(ctx, searchTerm)
	if err != nil {
		r.logger.Error("failed to get total count of search products", zap.Error(err))
		return model.PaginationResult[[]model.SearchProductResult]{}, fmt.Errorf("count search products: %w", err)
	}

	totalPages := int32((count + int64(limit) - 1) / int64(limit))
	pagination := model.PaginationResult[[]model.SearchProductResult]{
		TotalCount: count,
		TotalPages: totalPages,
		Page:       int32(page),
		PageSize:   int32(pageSize),
		Data:       results,
	}
	return pagination, nil
}
