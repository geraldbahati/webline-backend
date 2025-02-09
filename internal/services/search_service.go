package services

import (
	"context"
	"time"

	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"

	"go.uber.org/zap"
)

// TTL for caching search results (adjust as needed).
const searchCacheTTL = 5 * time.Minute

// SearchService provides search functionality with an optimal caching layer.
type SearchService struct {
	searchRepository repository.SearchRepository
	cacheService     CacheService
	logger           *zap.Logger
}

// NewSearchService creates a new instance of SearchService.
func NewSearchService(repo repository.SearchRepository, cache CacheService, logger *zap.Logger) *SearchService {
	return &SearchService{
		searchRepository: repo,
		cacheService:     cache,
		logger:           logger,
	}
}

// SearchProducts performs a search query and returns matching products.
// For non‑paginated queries we use page 1 by default.
func (s *SearchService) SearchProducts(ctx context.Context, searchTerm string, limit int) ([]model.SearchProductResult, error) {
	const defaultPage = 1
	cacheKey := SearchProductsKey(searchTerm, defaultPage, limit)
	var results []model.SearchProductResult

	err := s.cacheService.GetOrSetWithTTL(ctx, cacheKey, &results, func() error {
		res, err := s.searchRepository.V2SearchProducts(ctx, searchTerm, limit)
		if err != nil {
			return err
		}
		results = res
		return nil
	}, searchCacheTTL)
	if err != nil {
		s.logger.Warn("Cache GetOrSetWithTTL error for search products, fetching directly", zap.Error(err))
		return s.searchRepository.V2SearchProducts(ctx, searchTerm, limit)
	}
	return results, nil
}

// SearchProductsPaginated performs a paginated search query and returns paginated results.
func (s *SearchService) SearchProductsPaginated(ctx context.Context, searchTerm string, page int, pageSize int) (model.PaginationResult[[]model.SearchProductResult], error) {
	cacheKey := SearchProductsKey(searchTerm, page, pageSize)
	var paginationResult model.PaginationResult[[]model.SearchProductResult]

	err := s.cacheService.GetOrSetWithTTL(ctx, cacheKey, &paginationResult, func() error {
		result, err := s.searchRepository.V2SearchProductsPaginated(ctx, searchTerm, page, pageSize)
		if err != nil {
			return err
		}
		paginationResult = result
		return nil
	}, searchCacheTTL)
	if err != nil {
		s.logger.Warn("Cache GetOrSetWithTTL error for paginated search, fetching directly", zap.Error(err))
		return s.searchRepository.V2SearchProductsPaginated(ctx, searchTerm, page, pageSize)
	}
	return paginationResult, nil
}

// AutocompleteSuggestions returns autocomplete search suggestions for the given search term.
func (s *SearchService) AutocompleteSuggestions(ctx context.Context, searchTerm string, limit int) ([]string, error) {
	cacheKey := AutocompleteSuggestionsKey(searchTerm, limit)
	var suggestions []string

	err := s.cacheService.GetOrSetWithTTL(ctx, cacheKey, &suggestions, func() error {
		res, err := s.searchRepository.AutocompleteSuggestions(ctx, searchTerm, limit)
		if err != nil {
			return err
		}
		suggestions = res
		return nil
	}, searchCacheTTL)
	if err != nil {
		s.logger.Warn("Cache GetOrSetWithTTL error for autocomplete suggestions, fetching directly", zap.Error(err))
		return s.searchRepository.AutocompleteSuggestions(ctx, searchTerm, limit)
	}
	return suggestions, nil
}
