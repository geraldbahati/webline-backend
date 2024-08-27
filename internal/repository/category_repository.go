package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sort"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"

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
func (r *CategoryRepository) execTx(ctx context.Context, fn func(*database.Queries) error) (err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	q := database.New(tx)
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			r.logger.Error("transaction failed, rolling back", zap.Error(err))
			if rbErr := tx.Rollback(); rbErr != nil {
				r.logger.Error("rollback failed", zap.Error(rbErr))
				err = fmt.Errorf("rollback transaction: %w", rbErr)
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				err = fmt.Errorf("commit transaction: %w", commitErr)
			}
		}
	}()

	err = fn(q)
	return err
}

// CreateCategory stores a category in the database and returns the created category
func (r *CategoryRepository) CreateCategory(
	ctx context.Context,
	category database.CreateCategoryParams,
) error {
	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		err = q.CreateCategory(ctx, category)
		if err != nil {
			return fmt.Errorf("failed to create category: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("failed to create category", zap.Error(err))
		return err
	}
	return nil
}

// GetCategoryByID retrieves a category by its ID
func (r *CategoryRepository) GetCategoryByID(
	ctx context.Context,
	id uuid.UUID,
) (database.GetCategoryByIDRow, error) {
	category, err := r.Queries.GetCategoryByID(ctx, id)
	if err != nil {
		r.logger.Error("failed to get category by ID", zap.Error(err))
		return database.GetCategoryByIDRow{}, fmt.Errorf("failed to get category by ID: %w", err)
	}
	return category, nil
}

// GetCategories retrieves all categories
func (r *CategoryRepository) GetCategories(ctx context.Context) ([]database.ListCategoriesRow, error) {
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
	params database.UpdateCategoryParams,
) error {

	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		_, err = q.UpdateCategory(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to update category: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("failed to update category", zap.Error(err))
		return err
	}
	return nil
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
) ([]database.GetCategoriesByParentIDRow, error) {
	log.Printf("parentID: %v", parentID)
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
) ([]database.GetParentCategoriesRow, error) {
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
) (database.GetCategoryByNameRow, error) {
	category, err := r.Queries.GetCategoryByName(ctx, name)
	if err != nil {
		r.logger.Error("failed to get category by name", zap.Error(err))
		return database.GetCategoryByNameRow{}, fmt.Errorf("failed to get category by name: %w", err)
	}
	return category, nil
}

// UpdateCategoryImage updates the image of a category
func (r *CategoryRepository) UpdateCategoryImage(
	ctx context.Context,
	id uuid.UUID,
	imageURL string,
) (database.UpdateCategoryImageRow, error) {
	var updatedCategory database.UpdateCategoryImageRow
	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		updatedCategory, err = q.UpdateCategoryImage(ctx, database.UpdateCategoryImageParams{
			ID:       id,
			ImageUrl: sql.NullString{String: imageURL, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("failed to update category image: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("failed to update category image", zap.Error(err))
		return database.UpdateCategoryImageRow{}, err
	}

	r.logger.Info("Category image successfully updated")
	return updatedCategory, nil
}

// GetV2CategoryHierarchy retrieves the category hierarchy for the V2 API
func (r *CategoryRepository) GetV2CategoryHierarchy(
	ctx context.Context,
) ([]*model.V2CategoryHierarchy, error) {
	rows, err := r.Queries.GetV2CategoryHierarchy(ctx)
	if err != nil {
		r.logger.Error("failed to get V2 category hierarchy", zap.Error(err))
		return nil, fmt.Errorf("failed to get V2 category hierarchy: %w", err)
	}

	categoryMap := make(map[string]*model.V2CategoryHierarchy)
	var rootCategories []*model.V2CategoryHierarchy

	for _, row := range rows {
		category := &model.V2CategoryHierarchy{
			ID:               row.ID.String(),
			Name:             row.Name,
			Position:         int(row.Position),
			IsActive:         row.IsActive,
			NumberOfProducts: int(row.TotalProducts),
			Slug:             row.Slug,
		}

		categoryMap[category.ID] = category

		if row.ParentID.Valid {
			parentCategory, exists := categoryMap[row.ParentID.UUID.String()]
			if exists {
				parentCategory.Children = append(parentCategory.Children, category)
				sort.SliceStable(parentCategory.Children, func(i, j int) bool {
					return parentCategory.Children[i].Position < parentCategory.Children[j].Position
				})
			}
		} else {
			rootCategories = append(rootCategories, category)
		}
	}

	sort.SliceStable(rootCategories, func(i, j int) bool {
		return rootCategories[i].Position < rootCategories[j].Position
	})

	return rootCategories, nil

}

// DeleteCategory deletes a category from the database
func (r *CategoryRepository) DeleteCategory(
	ctx context.Context,
	id uuid.UUID,
) error {
	err := r.execTx(ctx, func(q *database.Queries) error {
		err := q.HardDeleteCategory(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to delete category: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("failed to delete category", zap.Error(err))
		return err
	}

	r.logger.Info("Category successfully deleted")
	return nil
}

// GetCategoryBySlug retrieves a category by its slug
func (r *CategoryRepository) GetCategoryBySlug(
	ctx context.Context,
	slug string,
) (*uuid.UUID, error) {
	id, err := r.Queries.GetCategoryBySlug(ctx, slug)
	if err != nil {
		r.logger.Error("failed to get category by slug", zap.Error(err))
		return nil, fmt.Errorf("failed to get category by slug: %w", err)
	}
	return &id, nil
}

// GetCategoryDetailsBySlug retrieves a category detail by its slug
func (r *CategoryRepository) GetCategoryDetailsBySlug(
	ctx context.Context,
	slug string,
) (*model.V2CategoryDetail, error) {
	category, err := r.Queries.GetCategoryDetailsBySlug(ctx, slug)
	if err != nil {
		r.logger.Error("failed to get category detail by slug", zap.Error(err))
		return nil, fmt.Errorf("failed to get category detail by slug: %w", err)
	}

	parentID := ""
	if category.ParentID.Valid {
		parentID = category.ParentID.UUID.String()
	}

	return &model.V2CategoryDetail{
		Slug:            category.Slug,
		Name:            category.Name,
		Description:     category.Description.String,
		ParentID:        parentID,
		MetaTitle:       category.MetaTitle.String,
		MetaDescription: category.MetaDescription.String,
		ImageURL:        category.ImageUrl.String,
	}, nil

}

// GetCategorySEOBySlug retrieves a category SEO by its slug
func (r *CategoryRepository) GetCategorySEOBySlug(
	ctx context.Context,
	slug string,
) (*model.CategorySEO, error) {
	seo, err := r.Queries.GetCategorySEOBySlug(ctx, slug)
	if err != nil {
		r.logger.Error("failed to get category SEO by slug", zap.Error(err))
		return nil, fmt.Errorf("failed to get category SEO by slug: %w", err)
	}

	return &model.CategorySEO{
		MetaTitle:       seo.MetaTitle.String,
		MetaDescription: seo.MetaDescription.String,
	}, nil
}
