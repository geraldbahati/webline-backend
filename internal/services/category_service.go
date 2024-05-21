package services

import (
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"weblineBackend/internal/database"
	"weblineBackend/internal/repository"
)

type CategoryService struct {
	repo   *repository.CategoryRepository
	logger *zap.Logger
}

func NewCategoryService(repo *repository.CategoryRepository, logger *zap.Logger) *CategoryService {
	return &CategoryService{
		repo:   repo,
		logger: logger,
	}
}

// CreateCategoryService creates a new category
func (s *CategoryService) CreateCategoryService(
	ctx context.Context,
	name string,
	parentID string,
) (database.Category, error) {

	// parse parentID to null uuid
	var parentIDValue uuid.NullUUID
	if parentID != "" {
		id, err := uuid.Parse(parentID)
		if err != nil {
			s.logger.Error("failed to parse parent ID", zap.Error(err))
			return database.Category{}, err
		}
		parentIDValue = uuid.NullUUID{
			UUID:  id,
			Valid: true,
		}
	} else {
		parentIDValue = uuid.NullUUID{
			UUID:  uuid.Nil,
			Valid: false,
		}
	}

	categoryParams := database.CreateCategoryParams{
		Name:     name,
		ParentID: parentIDValue,
	}

	category, err := s.repo.CreateCategory(ctx, categoryParams)
	if err != nil {
		s.logger.Error("failed to create category", zap.Error(err))
		return database.Category{}, err
	}

	return category, nil
}

// GetCategoryByIDService retrieves a category by its ID
func (s *CategoryService) GetCategoryByIDService(ctx context.Context, id string) (database.Category, error) {
	// parse id to uuid
	categoryID, err := uuid.Parse(id)

	category, err := s.repo.GetCategoryByID(ctx, categoryID)
	if err != nil {
		s.logger.Error("failed to get category by ID", zap.Error(err))
		return database.Category{}, err
	}

	return category, nil
}

// GetCategoriesService retrieves all categories
func (s *CategoryService) GetCategoriesService(ctx context.Context) ([]database.Category, error) {
	categories, err := s.repo.GetCategories(ctx)
	if err != nil {
		s.logger.Error("failed to get categories", zap.Error(err))
		return nil, err
	}

	return categories, nil
}

// UpdateCategoryService updates a category in the database and returns the updated category
func (s *CategoryService) UpdateCategoryService(
	ctx context.Context,
	id string,
	name string,
	parentID string,
) (database.Category, error) {
	// parse id to uuid
	categoryID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("failed to parse category ID", zap.Error(err))
		return database.Category{}, err
	}

	// parse parentID to null uuid
	var parentIDValue uuid.NullUUID
	if parentID != "" {
		id, err := uuid.Parse(parentID)
		if err != nil {
			s.logger.Error("failed to parse parent ID", zap.Error(err))
			return database.Category{}, err
		}
		parentIDValue = uuid.NullUUID{
			UUID:  id,
			Valid: true,
		}
	} else {
		parentIDValue = uuid.NullUUID{
			UUID:  uuid.Nil,
			Valid: false,
		}
	}

	category, err := s.repo.UpdateCategory(ctx, categoryID, name, parentIDValue)
	if err != nil {
		s.logger.Error("failed to update category", zap.Error(err))
		return database.Category{}, err
	}

	return category, nil
}

// SoftDeleteCategoryService marks a category as inactive
func (s *CategoryService) SoftDeleteCategoryService(ctx context.Context, id string) error {
	// parse id to uuid
	categoryID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("failed to parse category ID", zap.Error(err))
		return err
	}

	err = s.repo.SoftDeleteCategory(ctx, categoryID)
	if err != nil {
		s.logger.Error("failed to soft delete category", zap.Error(err))
		return err
	}

	return nil
}

// GetCategoriesByParentIDService retrieves categories by their parent ID
func (s *CategoryService) GetCategoriesByParentIDService(ctx context.Context, parentID string) ([]database.Category, error) {
	// parse parentID to null uuid
	var parentIDValue uuid.NullUUID
	if parentID != "" {
		id, err := uuid.Parse(parentID)
		if err != nil {
			s.logger.Error("failed to parse parent ID", zap.Error(err))
			return nil, err
		}
		parentIDValue = uuid.NullUUID{
			UUID:  id,
			Valid: true,
		}
	} else {
		parentIDValue = uuid.NullUUID{
			UUID:  uuid.Nil,
			Valid: false,
		}
	}

	categories, err := s.repo.GetCategoriesByParentID(ctx, parentIDValue)
	if err != nil {
		s.logger.Error("failed to get categories by parent ID", zap.Error(err))
		return nil, err
	}

	return categories, nil
}

// GetCategoriesWithProductsCountService retrieves categories along with the count of products in each category
func (s *CategoryService) GetCategoriesWithProductsCountService(ctx context.Context) ([]database.GetCategoriesWithProductsCountRow, error) {
	categoriesWithCount, err := s.repo.GetCategoriesWithProductsCount(ctx)
	if err != nil {
		s.logger.Error("failed to get categories with products count", zap.Error(err))
		return nil, err
	}

	return categoriesWithCount, nil
}

// GetCategoryTreeService retrieves the full category hierarchy
func (s *CategoryService) GetCategoryTreeService(ctx context.Context) ([]database.GetCategoryTreeRow, error) {
	categories, err := s.repo.GetCategoryTree(ctx)
	if err != nil {
		s.logger.Error("failed to get category tree", zap.Error(err))
		return nil, err
	}

	return categories, nil
}

// CheckCategoryExistenceService checks if a category exists in the database
func (s *CategoryService) CheckCategoryExistenceService(ctx context.Context, id string) (bool, error) {
	// parse id to uuid
	categoryID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("failed to parse category ID", zap.Error(err))
		return false, err
	}

	exists, err := s.repo.CheckCategoryExistence(ctx, categoryID)
	if err != nil {
		s.logger.Error("failed to check category existence", zap.Error(err))
		return false, err
	}

	return exists, nil
}

// GetCategoriesWithSubcategoryCountService retrieves categories along with the count of subcategories for each category
func (s *CategoryService) GetCategoriesWithSubcategoryCountService(ctx context.Context) ([]database.GetCategoriesWithSubcategoryCountRow, error) {
	categoriesWithCount, err := s.repo.GetCategoriesWithSubcategoryCount(ctx)
	if err != nil {
		s.logger.Error("failed to get categories with subcategory count", zap.Error(err))
		return nil, err
	}

	return categoriesWithCount, nil
}
