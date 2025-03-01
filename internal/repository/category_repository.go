package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sort"
	"time"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"

	"errors"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	// Constants for retry mechanism
	maxRetries     = 3
	initialBackoff = 100 * time.Millisecond
	queryTimeout   = 30 * time.Second
	// Error codes
	pqDeadlockCode  = "40P01"
	pqSerializeCode = "40001"
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

// execTx executes a database transaction with the provided function and retry mechanism
func (r *CategoryRepository) execTx(ctx context.Context, fn func(*database.Queries) error) (err error) {
	var retryCount int
	backoff := initialBackoff
	opStart := time.Now()

	for retryCount <= maxRetries {
		// Create a transaction with timeout
		txCtx, cancel := context.WithTimeout(ctx, queryTimeout)
		tx, err := r.db.BeginTx(txCtx, &sql.TxOptions{Isolation: sql.LevelSerializable})

		if err != nil {
			cancel()
			r.logger.Error("failed to begin transaction",
				zap.Error(err),
				zap.Int("retry", retryCount),
				zap.Duration("backoff", backoff))

			// Check for context cancellation
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return fmt.Errorf("operation canceled or timed out: %w", err)
			}

			// Handle database connection issues
			if retryCount < maxRetries {
				retryCount++
				time.Sleep(backoff)
				backoff *= 2 // Exponential backoff
				continue
			}
			return fmt.Errorf("begin transaction: %w", err)
		}

		q := database.New(tx)
		txErr := fn(q)

		if txErr == nil {
			// Log slow transactions if they take too long
			opDuration := time.Since(opStart)
			if opDuration > 500*time.Millisecond {
				r.logger.Warn("slow transaction detected",
					zap.Duration("duration", opDuration),
					zap.Int("retries", retryCount))
			}

			if commitErr := tx.Commit(); commitErr != nil {
				r.logger.Error("commit failed",
					zap.Error(commitErr),
					zap.Int("retry", retryCount),
					zap.Duration("duration", opDuration))
				cancel()

				// Check for context issues
				if errors.Is(commitErr, context.Canceled) || errors.Is(commitErr, context.DeadlineExceeded) {
					return fmt.Errorf("commit canceled or timed out: %w", commitErr)
				}

				// Retry on specific PostgreSQL errors like serialization failures
				if pqErr, ok := commitErr.(*pq.Error); ok {
					if pqErr.Code == pqSerializeCode && retryCount < maxRetries {
						retryCount++
						time.Sleep(backoff)
						backoff *= 2
						continue
					}
				}
				return fmt.Errorf("commit transaction: %w", commitErr)
			}

			// Success case
			cancel()
			return nil
		}

		// Transaction failed, evaluate if we should retry
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			r.logger.Error("rollback failed after transaction error",
				zap.Error(rollbackErr),
				zap.Error(txErr))
		}

		cancel()

		// Check for context-related issues
		if errors.Is(txErr, context.Canceled) || errors.Is(txErr, context.DeadlineExceeded) {
			return fmt.Errorf("operation canceled or timed out during execution: %w", txErr)
		}

		// Check if error is retryable (deadlock, serialization failure)
		if pqErr, ok := txErr.(*pq.Error); ok {
			if (pqErr.Code == pqDeadlockCode || pqErr.Code == pqSerializeCode) && retryCount < maxRetries {
				r.logger.Warn("transaction failed with retryable error",
					zap.String("code", string(pqErr.Code)),
					zap.Error(txErr),
					zap.Int("retry", retryCount))
				retryCount++
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
		}

		// Non-retryable error
		r.logger.Error("transaction failed permanently",
			zap.Error(txErr),
			zap.Int("retries", retryCount),
			zap.Duration("elapsed", time.Since(opStart)))
		return txErr
	}

	return fmt.Errorf("transaction failed after %d retries (total time: %v)",
		maxRetries, time.Since(opStart))
}

// withTimeout adds a timeout to the context for database operations
func (r *CategoryRepository) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, queryTimeout)
}

// CreateCategory stores a category in the database with optimized error handling
func (r *CategoryRepository) CreateCategory(
	ctx context.Context,
	category database.CreateCategoryParams,
) error {
	// Add slug validation before proceeding
	if category.Name == "" {
		return fmt.Errorf("category name cannot be empty")
	}

	err := r.execTx(ctx, func(q *database.Queries) error {
		err := q.CreateCategory(ctx, category)
		if err != nil {
			// Check for specific error cases like unique constraint violations
			if pqErr, ok := err.(*pq.Error); ok {
				if pqErr.Code == "23505" { // unique violation
					return fmt.Errorf("category with this name already exists")
				}
			}
			return fmt.Errorf("failed to create category: %w", err)
		}
		return nil
	})

	if err != nil {
		r.logger.Error("failed to create category",
			zap.Error(err),
			zap.String("name", category.Name))
		return err
	}

	r.logger.Info("category created successfully", zap.String("name", category.Name))
	return nil
}

// GetCategoryByID retrieves a category by its ID with timeout
func (r *CategoryRepository) GetCategoryByID(
	ctx context.Context,
	id uuid.UUID,
) (database.GetCategoryByIDRow, error) {
	ctxWithTimeout, cancel := r.withTimeout(ctx)
	defer cancel()

	category, err := r.Queries.GetCategoryByID(ctxWithTimeout, id)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Debug("category not found", zap.String("id", id.String()))
			return database.GetCategoryByIDRow{}, fmt.Errorf("category not found: %w", err)
		}
		r.logger.Error("failed to get category by ID", zap.Error(err), zap.String("id", id.String()))
		return database.GetCategoryByIDRow{}, fmt.Errorf("failed to get category by ID: %w", err)
	}

	return category, nil
}

// GetCategories retrieves all categories with efficient batch processing
func (r *CategoryRepository) GetCategories(ctx context.Context, activeOnly bool) ([]database.ListCategoriesRow, error) {
	ctxWithTimeout, cancel := r.withTimeout(ctx)
	defer cancel()

	categories, err := r.Queries.ListCategories(ctxWithTimeout, activeOnly)
	if err != nil {
		r.logger.Error("failed to get categories",
			zap.Error(err),
			zap.Bool("activeOnly", activeOnly))
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	r.logger.Debug("categories successfully retrieved",
		zap.Int("count", len(categories)),
		zap.Bool("activeOnly", activeOnly))
	return categories, nil
}

// BulkCreateCategories creates multiple categories in a single transaction
func (r *CategoryRepository) BulkCreateCategories(
	ctx context.Context,
	categories []database.CreateCategoryParams,
) error {
	if len(categories) == 0 {
		return nil
	}

	err := r.execTx(ctx, func(q *database.Queries) error {
		for _, category := range categories {
			if err := q.CreateCategory(ctx, category); err != nil {
				return fmt.Errorf("failed to create category '%s': %w", category.Name, err)
			}
		}
		return nil
	})

	if err != nil {
		r.logger.Error("failed to bulk create categories",
			zap.Error(err),
			zap.Int("count", len(categories)))
		return err
	}

	r.logger.Info("bulk category creation completed", zap.Int("count", len(categories)))
	return nil
}

// SearchCategoriesWithAncestors searches for categories and includes their ancestry path
func (r *CategoryRepository) SearchCategoriesWithAncestors(
	ctx context.Context,
	query string,
	activeOnly bool,
	limit int32,
	offset int32,
) ([]model.CategoryWithAncestry, error) {
	// First get matching categories
	params := database.SearchCategoriesParams{
		Column1: sql.NullString{String: fmt.Sprintf("%%%s%%", query), Valid: true}, // Use % for LIKE pattern
		Column2: activeOnly,
		Limit:   limit,
		Offset:  offset,
	}

	ctxWithTimeout, cancel := r.withTimeout(ctx)
	defer cancel()

	searchResults, err := r.Queries.SearchCategories(ctxWithTimeout, params)
	if err != nil {
		r.logger.Error("failed to search categories",
			zap.Error(err),
			zap.String("query", query))
		return nil, fmt.Errorf("failed to search categories: %w", err)
	}

	if len(searchResults) == 0 {
		return []model.CategoryWithAncestry{}, nil
	}

	// Get ancestors for each matching category
	result := make([]model.CategoryWithAncestry, 0, len(searchResults))

	// Use a map to deduplicate parent fetching
	categoryIDs := make([]uuid.UUID, 0, len(searchResults))
	for _, cat := range searchResults {
		categoryIDs = append(categoryIDs, cat.ID)
	}

	// For all found categories, get their ancestry paths in bulk
	ancestryPaths := make(map[string][]database.GetCategoryAncestorsRow)

	// Execute in batches to avoid excessive queries
	batchSize := 10
	for i := 0; i < len(categoryIDs); i += batchSize {
		end := i + batchSize
		if end > len(categoryIDs) {
			end = len(categoryIDs)
		}

		batch := categoryIDs[i:end]
		for _, id := range batch {
			params := database.GetCategoryAncestorsParams{
				ID:      id,
				Column2: activeOnly,
			}

			ancestors, err := r.Queries.GetCategoryAncestors(ctx, params)
			if err != nil {
				r.logger.Warn("failed to get ancestors for category",
					zap.Error(err),
					zap.String("id", id.String()))
				continue
			}

			ancestryPaths[id.String()] = ancestors
		}
	}

	// Combine search results with ancestry paths
	for _, cat := range searchResults {
		ancestry := model.CategoryWithAncestry{
			ID:           cat.ID.String(),
			Name:         cat.Name,
			Slug:         cat.Slug,
			Description:  cat.Description.String,
			IsActive:     cat.IsActive,
			ProductCount: 0, // Will be populated if available
			AncestryPath: []model.CategoryBreadcrumb{},
		}

		// Add ancestors if available
		if ancestors, ok := ancestryPaths[cat.ID.String()]; ok {
			for _, ancestor := range ancestors {
				ancestry.AncestryPath = append(ancestry.AncestryPath, model.CategoryBreadcrumb{
					ID:   ancestor.ID.String(),
					Name: ancestor.Name,
					Slug: ancestor.Slug,
				})
			}
		}

		result = append(result, ancestry)
	}

	r.logger.Debug("search with ancestry completed",
		zap.String("query", query),
		zap.Int("results", len(result)))

	return result, nil
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
	activeOnly bool,
) ([]*model.V2CategoryHierarchy, error) {
	rows, err := r.Queries.GetV2CategoryHierarchy(ctx, activeOnly)
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

// GetDirectChildrenCategories retrieves immediate children of a category
func (r *CategoryRepository) GetDirectChildrenCategories(
	ctx context.Context,
	parentID uuid.UUID,
	activeOnly bool,
) ([]database.GetDirectChildrenCategoriesRow, error) {
	params := database.GetDirectChildrenCategoriesParams{
		ParentID: uuid.NullUUID{UUID: parentID, Valid: true},
		Column2:  activeOnly,
	}

	categories, err := r.Queries.GetDirectChildrenCategories(ctx, params)
	if err != nil {
		r.logger.Error("failed to get direct children categories",
			zap.Error(err),
			zap.String("parent_id", parentID.String()))
		return nil, fmt.Errorf("failed to get direct children categories: %w", err)
	}

	r.logger.Debug("direct children categories retrieved successfully",
		zap.Int("count", len(categories)),
		zap.String("parent_id", parentID.String()))
	return categories, nil
}

// GetAllChildrenCategories retrieves all descendants of a category
func (r *CategoryRepository) GetAllChildrenCategories(
	ctx context.Context,
	parentID uuid.UUID,
	activeOnly bool,
) ([]database.GetAllChildrenCategoriesRow, error) {
	params := database.GetAllChildrenCategoriesParams{
		ID:      parentID,
		Column2: activeOnly,
	}

	categories, err := r.Queries.GetAllChildrenCategories(ctx, params)
	if err != nil {
		r.logger.Error("failed to get all children categories",
			zap.Error(err),
			zap.String("parent_id", parentID.String()))
		return nil, fmt.Errorf("failed to get all children categories: %w", err)
	}

	r.logger.Debug("all children categories retrieved successfully",
		zap.Int("count", len(categories)),
		zap.String("parent_id", parentID.String()))
	return categories, nil
}

// GetAllChildrenCategoriesWithProductCount retrieves all descendants with product counts
func (r *CategoryRepository) GetAllChildrenCategoriesWithProductCount(
	ctx context.Context,
	parentID uuid.UUID,
	activeOnly bool,
) ([]database.GetAllChildrenCategoriesWithProductCountRow, error) {
	params := database.GetAllChildrenCategoriesWithProductCountParams{
		ID:      parentID,
		Column2: activeOnly,
	}

	categories, err := r.Queries.GetAllChildrenCategoriesWithProductCount(ctx, params)
	if err != nil {
		r.logger.Error("failed to get children categories with product count",
			zap.Error(err),
			zap.String("parent_id", parentID.String()))
		return nil, fmt.Errorf("failed to get children categories with product count: %w", err)
	}

	r.logger.Debug("children categories with product count retrieved successfully",
		zap.Int("count", len(categories)),
		zap.String("parent_id", parentID.String()))
	return categories, nil
}

// GetChildrenCategoriesByDepth retrieves categories at a specific depth level
func (r *CategoryRepository) GetChildrenCategoriesByDepth(
	ctx context.Context,
	parentID uuid.UUID,
	depth int32,
) ([]database.GetChildrenCategoriesByDepthRow, error) {
	params := database.GetChildrenCategoriesByDepthParams{
		ID:   parentID,
		Path: sql.NullString{}, // Path will be populated by the SQL query
	}

	categories, err := r.Queries.GetChildrenCategoriesByDepth(ctx, params)
	if err != nil {
		r.logger.Error("failed to get children categories by depth",
			zap.Error(err),
			zap.String("parent_id", parentID.String()),
			zap.Int32("depth", depth))
		return nil, fmt.Errorf("failed to get children categories by depth: %w", err)
	}

	r.logger.Debug("children categories by depth retrieved successfully",
		zap.Int("count", len(categories)),
		zap.String("parent_id", parentID.String()),
		zap.Int32("depth", depth))
	return categories, nil
}

// BulkUpdateCategoryPositions updates positions of multiple categories in one operation
func (r *CategoryRepository) BulkUpdateCategoryPositions(
	ctx context.Context,
	categoryIDs []uuid.UUID,
	positions []int32,
) error {
	if len(categoryIDs) != len(positions) {
		err := fmt.Errorf("mismatched array lengths: %d category IDs vs %d positions", len(categoryIDs), len(positions))
		r.logger.Error("failed to bulk update category positions", zap.Error(err))
		return err
	}

	params := database.BulkUpdateCategoryPositionsParams{
		Column1: categoryIDs,
		Column2: positions,
	}

	err := r.Queries.BulkUpdateCategoryPositions(ctx, params)
	if err != nil {
		r.logger.Error("failed to bulk update category positions",
			zap.Error(err),
			zap.Int("count", len(categoryIDs)))
		return fmt.Errorf("failed to bulk update category positions: %w", err)
	}

	r.logger.Info("category positions updated successfully",
		zap.Int("count", len(categoryIDs)))
	return nil
}

// GetCategoryAncestors retrieves all ancestors of a category from root to parent
func (r *CategoryRepository) GetCategoryAncestors(
	ctx context.Context,
	categoryID uuid.UUID,
	activeOnly bool,
) ([]database.GetCategoryAncestorsRow, error) {
	params := database.GetCategoryAncestorsParams{
		ID:      categoryID,
		Column2: activeOnly,
	}

	ancestors, err := r.Queries.GetCategoryAncestors(ctx, params)
	if err != nil {
		r.logger.Error("failed to get category ancestors",
			zap.Error(err),
			zap.String("category_id", categoryID.String()))
		return nil, fmt.Errorf("failed to get category ancestors: %w", err)
	}

	r.logger.Debug("category ancestors retrieved successfully",
		zap.Int("count", len(ancestors)),
		zap.String("category_id", categoryID.String()))
	return ancestors, nil
}

// SearchCategories searches for categories by name with pagination
func (r *CategoryRepository) SearchCategories(
	ctx context.Context,
	query string,
	activeOnly bool,
	limit int32,
	offset int32,
) ([]database.SearchCategoriesRow, error) {
	params := database.SearchCategoriesParams{
		Column1: sql.NullString{String: query, Valid: true},
		Column2: activeOnly,
		Limit:   limit,
		Offset:  offset,
	}

	categories, err := r.Queries.SearchCategories(ctx, params)
	if err != nil {
		r.logger.Error("failed to search categories",
			zap.Error(err),
			zap.String("query", query),
			zap.Int32("limit", limit),
			zap.Int32("offset", offset))
		return nil, fmt.Errorf("failed to search categories: %w", err)
	}

	r.logger.Debug("categories search completed successfully",
		zap.Int("count", len(categories)),
		zap.String("query", query))
	return categories, nil
}

// GetCategoryStats modified for better error handling
func (r *CategoryRepository) GetCategoryStats(
	ctx context.Context,
	categoryID uuid.UUID,
) (database.GetCategoryStatsRow, error) {
	params := uuid.NullUUID{UUID: categoryID, Valid: true}
	ctxWithTimeout, cancel := r.withTimeout(ctx)
	defer cancel()

	stats, err := r.Queries.GetCategoryStats(ctxWithTimeout, params)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return empty stats instead of error for non-existent categories
			r.logger.Debug("no stats found for category", zap.String("category_id", categoryID.String()))
			return database.GetCategoryStatsRow{}, nil
		}

		r.logger.Error("failed to get category stats",
			zap.Error(err),
			zap.String("category_id", categoryID.String()))
		return database.GetCategoryStatsRow{}, fmt.Errorf("failed to get category stats: %w", err)
	}

	r.logger.Debug("category stats retrieved successfully",
		zap.String("category_id", categoryID.String()),
		zap.Int64("direct_children", stats.DirectChildrenCount),
		zap.Int64("all_descendants", stats.AllDescendantsCount),
		zap.Int64("direct_products", stats.DirectProductCount),
		zap.Int64("total_products", stats.TotalProductCount))
	return stats, nil
}

// GetPopularCategories retrieves categories with the most products
func (r *CategoryRepository) GetPopularCategories(
	ctx context.Context,
	limit int32,
) ([]database.GetPopularCategoriesRow, error) {
	categories, err := r.Queries.GetPopularCategories(ctx, limit)
	if err != nil {
		r.logger.Error("failed to get popular categories",
			zap.Error(err),
			zap.Int32("limit", limit))
		return nil, fmt.Errorf("failed to get popular categories: %w", err)
	}

	r.logger.Debug("popular categories retrieved successfully",
		zap.Int("count", len(categories)),
		zap.Int32("limit", limit))
	return categories, nil
}

// GetCategoryHierarchyStats optimized version
func (r *CategoryRepository) GetCategoryHierarchyStats(
	ctx context.Context,
	categoryID uuid.UUID,
	activeOnly bool,
) (*model.CategoryHierarchyStats, error) {
	// Create a supervisor context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	// Use errgroup to manage parallel queries and aggregate errors
	g, gCtx := errgroup.WithContext(ctxWithTimeout)

	// Create channels for results
	categoryChan := make(chan database.GetCategoryByIDRow, 1)
	statsChan := make(chan database.GetCategoryStatsRow, 1)
	childrenChan := make(chan []database.GetAllChildrenCategoriesWithProductCountRow, 1)

	// 1. Get the category details
	g.Go(func() error {
		category, err := r.Queries.GetCategoryByID(gCtx, categoryID)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("category not found with ID %s", categoryID)
			}
			return fmt.Errorf("failed to get category details: %w", err)
		}

		select {
		case categoryChan <- category:
			return nil
		case <-gCtx.Done():
			return gCtx.Err()
		}
	})

	// 2. Get category stats
	g.Go(func() error {
		statsParams := uuid.NullUUID{UUID: categoryID, Valid: true}
		stats, err := r.Queries.GetCategoryStats(gCtx, statsParams)
		if err != nil {
			if err == sql.ErrNoRows {
				// Return empty stats for non-existent category
				return nil
			}
			return fmt.Errorf("failed to get category stats: %w", err)
		}

		select {
		case statsChan <- stats:
			return nil
		case <-gCtx.Done():
			return gCtx.Err()
		}
	})

	// 3. Get children with product counts
	g.Go(func() error {
		childrenParams := database.GetAllChildrenCategoriesWithProductCountParams{
			ID:      categoryID,
			Column2: activeOnly,
		}

		children, err := r.Queries.GetAllChildrenCategoriesWithProductCount(gCtx, childrenParams)
		if err != nil {
			return fmt.Errorf("failed to get category children: %w", err)
		}

		select {
		case childrenChan <- children:
			return nil
		case <-gCtx.Done():
			return gCtx.Err()
		}
	})

	// Wait for all goroutines to complete and check for errors
	if err := g.Wait(); err != nil {
		r.logger.Error("error retrieving category hierarchy data",
			zap.Error(err),
			zap.String("category_id", categoryID.String()))
		return nil, err
	}

	// Get results from channels (no need to check closed channel since g.Wait() ensures completion)
	var category database.GetCategoryByIDRow
	var stats database.GetCategoryStatsRow
	var children []database.GetAllChildrenCategoriesWithProductCountRow

	// Get category
	select {
	case category = <-categoryChan:
	default:
		return nil, fmt.Errorf("missing category data")
	}

	// Get stats (use empty if not available)
	select {
	case stats = <-statsChan:
	default:
		// Use zero values if stats not available
	}

	// Get children (use empty slice if not available)
	select {
	case children = <-childrenChan:
	default:
		children = []database.GetAllChildrenCategoriesWithProductCountRow{}
	}

	// Build the response structure
	result := &model.CategoryHierarchyStats{
		Category: model.CategoryDetail{
			ID:              category.ID,
			Name:            category.Name,
			IsActive:        category.IsActive,
			ImageURL:        category.ImageUrl.String,
			SubCategories:   []model.CategoryDetail{},
			AvailableColors: []string{},
		},
		Stats: model.CategoryStats{
			DirectChildrenCount: int(stats.DirectChildrenCount),
			AllDescendantsCount: int(stats.AllDescendantsCount),
			DirectProductCount:  int(stats.DirectProductCount),
			TotalProductCount:   int(stats.TotalProductCount),
			DepthLevel:          int(stats.DepthLevel),
		},
		Children: []model.CategoryWithProductCount{},
	}

	if category.ParentID.Valid {
		result.Category.ParentID = category.ParentID.UUID
	}

	// Process children only if we have a reasonable number to avoid memory issues
	if len(children) > 500 {
		r.logger.Warn("large number of children categories",
			zap.Int("count", len(children)),
			zap.String("category_id", categoryID.String()))
	}

	// Convert children to model
	childrenMap := make(map[string][]model.CategoryWithProductCount)
	resultChildren := make([]model.CategoryWithProductCount, 0, len(children))

	for _, child := range children {
		childModel := model.CategoryWithProductCount{
			ID:           child.ID.String(),
			Name:         child.Name,
			Slug:         child.Slug,
			ImageURL:     child.ImageUrl.String,
			Description:  child.Description.String,
			IsActive:     child.IsActive,
			Position:     int(child.Position),
			Depth:        1, // Direct children are always at depth 1
			ProductCount: int(child.ProductCount),
		}

		if child.ParentID.Valid {
			parentID := child.ParentID.UUID.String()
			if parentID == categoryID.String() {
				resultChildren = append(resultChildren, childModel)
			} else {
				childrenMap[parentID] = append(childrenMap[parentID], childModel)
			}
		}
	}

	// Sort by position
	sort.SliceStable(resultChildren, func(i, j int) bool {
		return resultChildren[i].Position < resultChildren[j].Position
	})

	result.Children = resultChildren

	r.logger.Debug("category hierarchy stats retrieved successfully",
		zap.String("category_id", categoryID.String()),
		zap.Int("children_count", len(result.Children)),
		zap.Int("stats_total_products", result.Stats.TotalProductCount))

	return result, nil
}

// Add CategoryWithAncestry struct to model package
type CategoryWithAncestry struct {
	ID           string                     `json:"id"`
	Name         string                     `json:"name"`
	Slug         string                     `json:"slug"`
	Description  string                     `json:"description"`
	IsActive     bool                       `json:"isActive"`
	ProductCount int                        `json:"productCount"`
	AncestryPath []model.CategoryBreadcrumb `json:"ancestryPath"`
}

type CategoryBreadcrumb struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// GetDirectChildrenWithStats gets only immediate children with their stats in a single query
func (r *CategoryRepository) GetDirectChildrenWithStats(
	ctx context.Context,
	parentID uuid.UUID,
	activeOnly bool,
) ([]model.CategoryWithProductCount, error) {
	ctxWithTimeout, cancel := r.withTimeout(ctx)
	defer cancel()

	// Create params for the query
	params := database.GetDirectChildrenCategoriesParams{
		ParentID: uuid.NullUUID{UUID: parentID, Valid: true},
		Column2:  activeOnly,
	}

	// Get direct children
	directChildren, err := r.Queries.GetDirectChildrenCategories(ctxWithTimeout, params)
	if err != nil {
		r.logger.Error("failed to get direct children with stats",
			zap.Error(err),
			zap.String("parent_id", parentID.String()))
		return nil, fmt.Errorf("failed to get direct children: %w", err)
	}

	// Early return for empty results
	if len(directChildren) == 0 {
		return []model.CategoryWithProductCount{}, nil
	}

	// Prepare batch IDs for stats lookup
	childIDs := make([]uuid.UUID, len(directChildren))
	for i, child := range directChildren {
		childIDs[i] = child.ID
	}

	// Create a map to store product counts by category ID
	productCountMap := make(map[string]int)

	// Use a batch approach to get stats for all children efficiently
	batchSize := 25 // Adjust batch size based on your database capacity
	for i := 0; i < len(childIDs); i += batchSize {
		end := i + batchSize
		if end > len(childIDs) {
			end = len(childIDs)
		}

		batchIDs := childIDs[i:end]
		for _, id := range batchIDs {
			statsParams := uuid.NullUUID{UUID: id, Valid: true}

			// Get stats for this child
			stats, err := r.Queries.GetCategoryStats(ctxWithTimeout, statsParams)
			if err != nil {
				// Log error but continue with other categories
				r.logger.Warn("failed to get stats for child category",
					zap.Error(err),
					zap.String("child_id", id.String()))
				continue
			}

			// Log the product counts for debugging
			r.logger.Debug("category product counts",
				zap.String("category_id", id.String()),
				zap.Int64("direct_count", stats.DirectProductCount),
				zap.Int64("total_count", stats.TotalProductCount))

			// The TotalProductCount from GetCategoryStats includes products from this category and ALL subcategories
			productCountMap[id.String()] = int(stats.TotalProductCount)
		}
	}

	// Create result with the product counts
	result := make([]model.CategoryWithProductCount, 0, len(directChildren))
	for _, child := range directChildren {
		childID := child.ID.String()
		productCount := productCountMap[childID] // 0 if not found

		result = append(result, model.CategoryWithProductCount{
			ID:           childID,
			Name:         child.Name,
			Slug:         child.Slug,
			ImageURL:     child.ImageUrl.String,
			Description:  child.Description.String,
			IsActive:     child.IsActive,
			Position:     int(child.Position),
			Depth:        1, // Direct children are always at depth 1
			ProductCount: productCount,
		})
	}

	// Sort by position
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Position < result[j].Position
	})

	r.logger.Debug("direct children with stats retrieved successfully",
		zap.Int("count", len(result)),
		zap.String("parent_id", parentID.String()))
	return result, nil
}
