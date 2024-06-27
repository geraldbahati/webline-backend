package repository

import (
	"context"
	"database/sql"
	"fmt"
	"weblineBackend/internal/database"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CategoryRepository struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

// NewCategoryRepository initializes a new CategoryRepository with dependency injection for logging
func NewCategoryRepository(db *sql.DB, logger *zap.Logger) *CategoryRepository {
	return &CategoryRepository{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

// execTx executes a database transaction with the provided function
func (r *CategoryRepository) execTx(ctx context.Context, fn func(*database.Queries) error) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	q := database.New(tx)
	if err := fn(q); err != nil {
		r.logger.Error("transaction failed, rolling back", zap.Error(err))
		if rbErr := tx.Rollback(); rbErr != nil {
			r.logger.Error("rollback failed", zap.Error(rbErr))
			return fmt.Errorf("rollback transaction: %w", rbErr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// CreateCategory stores a category in the database and returns the created category
func (r *CategoryRepository) CreateCategory(
	ctx context.Context,
	category database.CreateCategoryParams,
) (database.Category, error) {
	var createdCategory database.Category
	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		createdCategory, err = q.CreateCategory(ctx, category)
		if err != nil {
			return fmt.Errorf("failed to create category: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("failed to create category", zap.Error(err))
		return database.Category{}, err
	}
	return createdCategory, nil
}

// GetCategoryByID retrieves a category by its ID
func (r *CategoryRepository) GetCategoryByID(
	ctx context.Context,
	id uuid.UUID,
) (database.Category, error) {
	category, err := r.Queries.GetCategoryByID(ctx, id)
	if err != nil {
		r.logger.Error("failed to get category by ID", zap.Error(err))
		return database.Category{}, fmt.Errorf("failed to get category by ID: %w", err)
	}
	return category, nil
}

// GetCategories retrieves all categories
func (r *CategoryRepository) GetCategories(ctx context.Context) ([]database.Category, error) {
	categories, err := r.Queries.ListCategories(ctx)
	if err != nil {
		r.logger.Error("failed to get categories", zap.Error(err))
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	r.logger.Info("Category successfully retrieved")
	return categories, nil
}

// UpdateCategory updates a category in the database and returns the updated category
func (r *CategoryRepository) UpdateCategory(
	ctx context.Context,
	id uuid.UUID,
	name string,
	parentID uuid.NullUUID,
	position int32,
) (database.Category, error) {
	var updatedCategory database.Category
	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		updatedCategory, err = q.UpdateCategory(ctx, database.UpdateCategoryParams{
			ID:       id,
			Name:     name,
			ParentID: parentID,
			Position: position,
		})
		if err != nil {
			return fmt.Errorf("failed to update category: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("failed to update category", zap.Error(err))
		return database.Category{}, err
	}
	return updatedCategory, nil
}

// SoftDeleteCategory marks a category as inactive in the database
func (r *CategoryRepository) SoftDeleteCategory(
	ctx context.Context,
	id uuid.UUID,
) error {
	err := r.Queries.SoftDeleteCategory(ctx, id)
	if err != nil {
		r.logger.Error("failed to soft delete category", zap.Error(err))
		return fmt.Errorf("failed to soft delete category: %w", err)
	}
	return nil
}

// GetCategoriesByParentID retrieves categories by their parent ID
func (r *CategoryRepository) GetCategoriesByParentID(
	ctx context.Context,
	parentID uuid.NullUUID,
) ([]database.Category, error) {
	categories, err := r.Queries.GetCategoriesByParentID(ctx, parentID)
	if err != nil {
		r.logger.Error("failed to get categories by parent ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get categories by parent ID: %w", err)
	}
	return categories, nil
}

// GetCategoriesWithProductsCount retrieves categories along with the count of products in each category
func (r *CategoryRepository) GetCategoriesWithProductsCount(
	ctx context.Context,
) ([]database.GetCategoriesWithProductsCountRow, error) {
	categoriesWithCount, err := r.Queries.GetCategoriesWithProductsCount(ctx)
	if err != nil {
		r.logger.Error("failed to get categories with products count", zap.Error(err))
		return nil, fmt.Errorf("failed to get categories with products count: %w", err)
	}
	return categoriesWithCount, nil
}

// GetCategoryTree retrieves the full category hierarchy
func (r *CategoryRepository) GetCategoryTree(
	ctx context.Context,
) ([]database.GetCategoryTreeRow, error) {
	categories, err := r.Queries.GetCategoryTree(ctx)
	if err != nil {
		r.logger.Error("failed to get category tree", zap.Error(err))
		return nil, fmt.Errorf("failed to get category tree: %w", err)
	}

	return categories, nil
}

// CheckCategoryExistence checks if a category exists in the database
func (r *CategoryRepository) CheckCategoryExistence(
	ctx context.Context,
	id uuid.UUID,
) (bool, error) {
	exists, err := r.Queries.CheckCategoryExistence(ctx, id)
	if err != nil {
		r.logger.Error("failed to check category existence", zap.Error(err))
		return false, fmt.Errorf("failed to check category existence: %w", err)
	}
	return exists, nil
}

// GetCategoriesWithSubcategoryCount retrieves categories along with the count of subcategories for each category
func (r *CategoryRepository) GetCategoriesWithSubcategoryCount(
	ctx context.Context,
) ([]database.GetCategoriesWithSubcategoryCountRow, error) {
	categoriesWithSubcategoryCount, err := r.Queries.GetCategoriesWithSubcategoryCount(ctx)
	if err != nil {
		r.logger.Error("failed to get categories with subcategory count", zap.Error(err))
		return nil, fmt.Errorf("failed to get categories with subcategory count: %w", err)
	}
	return categoriesWithSubcategoryCount, nil
}

// GetParentCategories retrieves parent categories
func (r *CategoryRepository) GetParentCategories(
	ctx context.Context,
) ([]database.Category, error) {
	parentCategories, err := r.Queries.GetParentCategories(ctx)
	if err != nil {
		r.logger.Error("failed to get parent categories", zap.Error(err))
		return nil, fmt.Errorf("failed to get parent categories: %w", err)
	}
	return parentCategories, nil
}

// GetCategoryByName retrieves a category by its name
func (r *CategoryRepository) GetCategoryByName(
	ctx context.Context,
	name string,
) (database.Category, error) {
	category, err := r.Queries.GetCategoryByName(ctx, name)
	if err != nil {
		r.logger.Error("failed to get category by name", zap.Error(err))
		return database.Category{}, fmt.Errorf("failed to get category by name: %w", err)
	}
	return category, nil
}

// GetCategoryHierarchy retrieves the category hierarchy
func (r *CategoryRepository) GetCategoryHierarchy(
	ctx context.Context,
) ([]database.GetCategoryHierarchyRow, error) {
	hierarchy, err := r.Queries.GetCategoryHierarchy(ctx)
	if err != nil {
		r.logger.Error("failed to get category hierarchy", zap.Error(err))
		return nil, fmt.Errorf("failed to get category hierarchy: %w", err)
	}

	r.logger.Info("Category hierarchy successfully retrieved")
	return hierarchy, nil
}
