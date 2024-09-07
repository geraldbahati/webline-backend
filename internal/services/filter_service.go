package services

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"log"
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"
	"weblineBackend/pkg/utils"
)

type FilterService struct {
	logger                    *zap.Logger
	filterCategoryProductRepo repository.FilterCategoryProductRepository
	filterProductRepo         repository.FilterProductRepository
	categoryRepo              *repository.CategoryRepository
	config                    *appconfig.Config
}

func NewFilterService(logger *zap.Logger, filterCategoryProductRepo repository.FilterCategoryProductRepository, filterProductRepo repository.FilterProductRepository, categoryRepo *repository.CategoryRepository, config *appconfig.Config) *FilterService {
	return &FilterService{
		logger:                    logger,
		filterCategoryProductRepo: filterCategoryProductRepo,
		filterProductRepo:         filterProductRepo,
		categoryRepo:              categoryRepo,
		config:                    config,
	}
}

// GetCategoryProductFilterOptions returns filter options for category products
func (s *FilterService) GetCategoryProductFilterOptions(ctx context.Context, categoryName string) (*model.FilterOptions, error) {
	// Get categoryID by categoryName
	category, err := s.categoryRepo.GetCategoryByName(ctx, categoryName)
	if err != nil {
		s.logger.Error("failed to get category by name", zap.Error(err))
		return nil, err
	}

	// Get product attributes and count by categoryID
	attributes, err := s.filterCategoryProductRepo.GetProductAttributesAndCountByCategoryID(ctx, category.ID)
	if err != nil {
		s.logger.Error("failed to get product attributes and count by categoryID", zap.Error(err))
		return nil, err
	}

	return attributes, nil
}

// GetProductFilterOptions returns filter options for products
func (s *FilterService) GetProductFilterOptions(ctx context.Context) (*model.FilterOptions, error) {
	// Get product attributes and count
	attributes, err := s.filterProductRepo.GetProductAttributes(ctx)
	if err != nil {
		s.logger.Error("failed to get product attributes and count", zap.Error(err))
		return nil, err
	}

	return attributes, nil
}

// GetProductsByFilters returns products by filters
func (s *FilterService) GetProductsByFilters(ctx context.Context, filterValues *model.AllProductFilterValues) (*model.PaginationResult[[]*model.Product], error) {
	// total count of products
	totalCount, err := s.filterProductRepo.GetTotalProductsByFilters(ctx, filterValues)
	if err != nil {
		s.logger.Error("failed to get total products by filters", zap.Error(err))
		return nil, err
	}

	log.Println("totalCount", totalCount)

	paginatedProducts, err := utils.Paginate(
		s.config,
		totalCount,
		filterValues.Offset,
		filterValues.Limit,
		func(offset int32, limit int32) ([]*model.Product, error) {
			// Update the offset and limit
			filterValues.Offset = offset
			filterValues.Limit = limit

			products, err := s.filterProductRepo.GetProductsByFilters(ctx, filterValues)
			if err != nil {
				s.logger.Error("failed to get products by filters", zap.Error(err))
				return nil, err
			}

			// Construct S3 URL for each product image
			for _, product := range products {
				if product.ImageURL != "" {
					product.ImageURL = s.constructS3URL(product.ImageURL)
				}
			}

			return products, nil
		},
	)
	if err != nil {
		s.logger.Error("failed to paginate products", zap.Error(err))
		return nil, err
	}

	return paginatedProducts, nil
}

// GetCategoryProductsByFilters returns category products by filters
func (s *FilterService) GetCategoryProductsByFilters(ctx context.Context, filterValues *model.CategoryProductFilterValues) (*model.PaginationResult[[]*model.Product], error) {
	// total count of category products
	totalCount, err := s.filterCategoryProductRepo.GetTotalCategoryProductsByFilters(ctx, filterValues)
	if err != nil {
		s.logger.Error("failed to get total category products by filters", zap.Error(err))
		return nil, err
	}

	paginatedProducts, err := utils.Paginate(
		s.config,
		totalCount,
		filterValues.Offset,
		filterValues.Limit,
		func(offset int32, limit int32) ([]*model.Product, error) {
			// Update the offset and limit
			filterValues.Offset = offset
			filterValues.Limit = limit

			products, err := s.filterCategoryProductRepo.GetCategoryProductsByFilters(ctx, filterValues)
			if err != nil {
				s.logger.Error("failed to get category products by filters", zap.Error(err))
				return nil, err
			}

			// Construct S3 URL for each product image
			for _, product := range products {
				if product.ImageURL != "" {
					product.ImageURL = s.constructS3URL(product.ImageURL)
				}
			}

			return products, nil
		},
	)
	if err != nil {
		s.logger.Error("failed to paginate category products", zap.Error(err))
		return nil, err
	}

	return paginatedProducts, nil
}

// constructS3URL constructs the S3 URL for a given file path
func (s *FilterService) constructS3URL(filePath string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.config.AWSBucketName, s.config.AWSRegion, filePath)
}
