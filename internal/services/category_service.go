package services

import (
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"log"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"
	"weblineBackend/internal/repository"
)

type CategoryService struct {
	categoryRepo     *repository.CategoryRepository
	productColorRepo *repository.ProductColourRepository
	logger           *zap.Logger
}

func NewCategoryService(categoryRepo *repository.CategoryRepository, productColorRepo *repository.ProductColourRepository, logger *zap.Logger) *CategoryService {
	return &CategoryService{
		categoryRepo:     categoryRepo,
		productColorRepo: productColorRepo,
		logger:           logger,
	}
}

// CreateCategoryService creates a new category
func (s *CategoryService) CreateCategoryService(
	ctx context.Context,
	name string,
	parentID string,
	position int32,
) (database.Category, error) {

	// parse parentID to null uuid
	log.Printf("parentID: %v", parentID)
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
		Position: position,
	}

	category, err := s.categoryRepo.CreateCategory(ctx, categoryParams)
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

	category, err := s.categoryRepo.GetCategoryByID(ctx, categoryID)
	if err != nil {
		s.logger.Error("failed to get category by ID", zap.Error(err))
		return database.Category{}, err
	}

	return category, nil
}

// GetCategoriesService retrieves all categories
func (s *CategoryService) GetCategoriesService(ctx context.Context) ([]database.Category, error) {
	categories, err := s.categoryRepo.GetCategories(ctx)
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
	position int32,
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

	category, err := s.categoryRepo.UpdateCategory(ctx, categoryID, name, parentIDValue, position)
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

	err = s.categoryRepo.SoftDeleteCategory(ctx, categoryID)
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

	categories, err := s.categoryRepo.GetCategoriesByParentID(ctx, parentIDValue)
	if err != nil {
		s.logger.Error("failed to get categories by parent ID", zap.Error(err))
		return nil, err
	}

	return categories, nil
}

// GetCategoriesWithProductsCountService retrieves categories along with the count of products in each category
func (s *CategoryService) GetCategoriesWithProductsCountService(ctx context.Context) ([]database.GetCategoriesWithProductsCountRow, error) {
	categoriesWithCount, err := s.categoryRepo.GetCategoriesWithProductsCount(ctx)
	if err != nil {
		s.logger.Error("failed to get categories with products count", zap.Error(err))
		return nil, err
	}

	return categoriesWithCount, nil
}

// GetCategoryTreeService retrieves the full category hierarchy
func (s *CategoryService) GetCategoryTreeService(ctx context.Context) ([]database.GetCategoryTreeRow, error) {
	categories, err := s.categoryRepo.GetCategoryTree(ctx)
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

	exists, err := s.categoryRepo.CheckCategoryExistence(ctx, categoryID)
	if err != nil {
		s.logger.Error("failed to check category existence", zap.Error(err))
		return false, err
	}

	return exists, nil
}

// GetCategoriesWithSubcategoryCountService retrieves categories along with the count of subcategories for each category
func (s *CategoryService) GetCategoriesWithSubcategoryCountService(ctx context.Context) ([]database.GetCategoriesWithSubcategoryCountRow, error) {
	categoriesWithCount, err := s.categoryRepo.GetCategoriesWithSubcategoryCount(ctx)
	if err != nil {
		s.logger.Error("failed to get categories with subcategory count", zap.Error(err))
		return nil, err
	}

	return categoriesWithCount, nil
}

// GetParentCategoriesService retrieves parent categories
func (s *CategoryService) GetParentCategoriesService(ctx context.Context) ([]database.Category, error) {
	categories, err := s.categoryRepo.GetParentCategories(ctx)
	if err != nil {
		s.logger.Error("failed to get parent categories", zap.Error(err))
		return nil, err
	}

	return categories, nil
}

// GetCategoryByNameService retrieves a category by its name
func (s *CategoryService) GetCategoryByNameService(ctx context.Context, name string) (model.Category, error) {
	// get category by name
	category, err := s.categoryRepo.GetCategoryByName(ctx, name)
	if err != nil {
		s.logger.Error("failed to get category by name", zap.Error(err))
		return model.Category{}, err
	}

	// get categories by parent ID
	categories, err := s.categoryRepo.GetCategoriesByParentID(ctx, uuid.NullUUID{UUID: category.ID, Valid: true})
	if err != nil {
		s.logger.Error("failed to get categories by parent ID", zap.Error(err))
		return model.Category{}, err
	}

	// get available colors for the category
	availableColors, err := s.productColorRepo.GetAvailableColorsByCategoryID(ctx, category.ID)
	if err != nil {
		s.logger.Error("failed to get available colors by category ID", zap.Error(err))
		return model.Category{}, err
	}

	// create a new category model
	var subCategories []model.Category
	for _, category := range categories {
		subCategory := model.Category{
			ID:       category.ID,
			Name:     category.Name,
			ParentID: category.ParentID.UUID,
			IsActive: category.IsActive,
		}
		subCategories = append(subCategories, subCategory)
	}

	categoryModel := model.Category{
		ID:              category.ID,
		Name:            category.Name,
		ParentID:        category.ParentID.UUID,
		IsActive:        category.IsActive,
		SubCategories:   subCategories,
		AvailableColors: availableColors,
	}

	return categoryModel, nil
}

// GetCategoryHierarchyService retrieves the category hierarchy
func (s *CategoryService) GetCategoryHierarchyService(ctx context.Context) ([]model.CategoryHierarchy, error) {
	// get category tree
	categories, err := s.categoryRepo.GetCategoryTree(ctx)
	if err != nil {
		s.logger.Error("failed to get category tree", zap.Error(err))
		return nil, err
	}

	// build the category hierarchy
	categoryHierarchy := buildCategoryHierarchy(categories)

	return categoryHierarchy, nil
}

func buildCategoryHierarchy(categories []database.GetCategoryTreeRow) []model.CategoryHierarchy {
	// create a map to store categories by their parent ID
	categoryMap := make(map[uuid.UUID][]database.GetCategoryTreeRow)
	for _, category := range categories {
		categoryMap[category.ParentID.UUID] = append(categoryMap[category.ParentID.UUID], category)
	}

	// create a slice to store the top-level categories
	var categoryHierarchy []model.CategoryHierarchy

	// iterate over the categories and build the hierarchy
	for _, category := range categoryMap[uuid.Nil] {
		categoryHierarchy = append(categoryHierarchy, model.CategoryHierarchy{
			Name:     category.Name,
			Children: buildCategoryHierarchyRecursively(category.ID, categoryMap),
		})
	}

	return categoryHierarchy
}

func buildCategoryHierarchyRecursively(parentID uuid.UUID, categoryMap map[uuid.UUID][]database.GetCategoryTreeRow) []model.CategoryHierarchy {
	// create a slice to store the children of the current category
	var children []model.CategoryHierarchy

	// iterate over the children of the current category
	for _, category := range categoryMap[parentID] {
		children = append(children, model.CategoryHierarchy{
			Name:     category.Name,
			Children: buildCategoryHierarchyRecursively(category.ID, categoryMap),
		})
	}

	return children
}
