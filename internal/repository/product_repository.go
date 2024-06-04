package repository

import (
	"context"
	"database/sql"
	"fmt"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"
	"weblineBackend/pkg/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ProductRepository struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

func NewProductRepository(db *sql.DB, logger *zap.Logger) *ProductRepository {
	return &ProductRepository{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

// execTx executes a database transaction with the provided function
func (r *ProductRepository) execTx(ctx context.Context, fn func(*database.Queries) error) error {
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

// CreateProduct stores a product in the database and returns the created product
func (r *ProductRepository) CreateProduct(
	ctx context.Context,
	product database.CreateProductParams,
) (model.ProductSchema, error) {
	var createdProduct model.ProductSchema
	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		product, err := q.CreateProduct(ctx, product)
		if err != nil {
			return fmt.Errorf("failed to create product: %w", err)
		}

		createdProduct = model.ProductSchema{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description.String,
			Price:       product.Price,
			Stock:       product.Stock.Int32,
			CategoryID:  product.CategoryID.UUID,
			IsActive:    product.IsActive.Bool,
			Featured:    product.Featured.Bool,
		}

		return nil
	})
	if err != nil {
		r.logger.Error("failed to create product", zap.Error(err))
		return model.ProductSchema{}, fmt.Errorf("failed to create product: %w", err)
	}
	return createdProduct, nil
}

// GetProductByID retrieves a product by its ID
func (r *ProductRepository) GetProductByID(
	ctx context.Context,
	id uuid.UUID,
) (model.ProductSchema, error) {
	product, err := r.Queries.GetProductByID(ctx, id)
	if err != nil {
		r.logger.Error("failed to get product by ID", zap.Error(err))
		return model.ProductSchema{}, fmt.Errorf("failed to get product by ID: %w", err)
	}

	return model.ProductSchema{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description.String,
		Price:       product.Price,
		Stock:       product.Stock.Int32,
		CategoryID:  product.CategoryID.UUID,
		IsActive:    product.IsActive.Bool,
		Featured:    product.Featured.Bool,
	}, nil
}

// ListProducts retrieves all products from the database
func (r *ProductRepository) ListProducts(ctx context.Context, limit int32, offset int32) ([]model.ProductSchema, error) {
	products, err := r.Queries.ListProducts(ctx, database.ListProductsParams{
		Limit:  limit,
		Offset: offset,
	})

	if err != nil {
		r.logger.Error("failed to list products", zap.Error(err))
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	var productSchemas []model.ProductSchema
	for _, product := range products {
		productSchemas = append(productSchemas, model.ProductSchema{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description.String,
			Price:       product.Price,
			Stock:       product.Stock.Int32,
			CategoryID:  product.CategoryID.UUID,
			IsActive:    product.IsActive.Bool,
			Featured:    product.Featured.Bool,
		})
	}

	return productSchemas, nil
}

// UpdateProduct updates a product in the database and returns the updated product
func (r *ProductRepository) UpdateProduct(
	ctx context.Context,
	params database.UpdateProductParams,
) (model.ProductSchema, error) {
	var updatedProduct model.ProductSchema
	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		product, err := q.UpdateProduct(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to update product: %w", err)
		}

		updatedProduct = model.ProductSchema{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description.String,
			Price:       product.Price,
			Stock:       product.Stock.Int32,
			CategoryID:  product.CategoryID.UUID,
			IsActive:    product.IsActive.Bool,
			Featured:    product.Featured.Bool,
		}

		return nil
	})
	if err != nil {
		r.logger.Error("failed to update product", zap.Error(err))
		return model.ProductSchema{}, fmt.Errorf("failed to update product: %w", err)
	}
	return updatedProduct, nil
}

// SoftDeleteProduct marks a product as inactive in the database
func (r *ProductRepository) SoftDeleteProduct(
	ctx context.Context,
	id uuid.UUID,
) error {
	err := r.execTx(ctx, func(q *database.Queries) error {
		if err := q.SoftDeleteProduct(ctx, id); err != nil {
			return fmt.Errorf("failed to soft delete product: %w", err)
		}
		return nil
	})
	if err != nil {
		r.logger.Error("failed to soft delete product", zap.Error(err))
		return fmt.Errorf("failed to soft delete product: %w", err)
	}
	return nil
}

// GetProductsByCategoryID retrieves products by their category ID
func (r *ProductRepository) GetProductsByCategoryID(
	ctx context.Context,
	categoryID uuid.UUID,
) ([]model.ProductSchema, error) {
	products, err := r.Queries.GetProductsByParentCategoryID(ctx, categoryID)
	if err != nil {
		r.logger.Error("failed to get products by category ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get products by category ID: %w", err)
	}

	var productSchemas []model.ProductSchema
	for _, product := range products {
		productSchemas = append(productSchemas, model.ProductSchema{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description.String,
			Price:       product.Price,
			Stock:       product.Stock.Int32,
			CategoryID:  product.CategoryID.UUID,
			IsActive:    product.IsActive.Bool,
			Featured:    product.Featured.Bool,
		})
	}

	return productSchemas, nil
}

// SearchProducts searches for products by name or description
func (r *ProductRepository) SearchProducts(
	ctx context.Context,
	searchTerm string,
) ([]model.ProductSchema, error) {
	products, err := r.Queries.SearchProducts(ctx, sql.NullString{String: searchTerm, Valid: true})
	if err != nil {
		r.logger.Error("failed to search products", zap.Error(err))
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	var productSchemas []model.ProductSchema
	for _, product := range products {
		productSchemas = append(productSchemas, model.ProductSchema{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description.String,
			Price:       product.Price,
			Stock:       product.Stock.Int32,
			CategoryID:  product.CategoryID.UUID,
			IsActive:    product.IsActive.Bool,
			Featured:    product.Featured.Bool,
		})
	}

	return productSchemas, nil
}

// CountProducts returns the total number of products in the database
func (r *ProductRepository) CountProducts(ctx context.Context) (int64, error) {
	count, err := r.Queries.CountProducts(ctx)
	if err != nil {
		r.logger.Error("failed to count products", zap.Error(err))
		return 0, fmt.Errorf("failed to count products: %w", err)
	}
	return count, nil
}

// GetProductsByParentCategoryID retrieves products by their parent category ID
func (r *ProductRepository) GetProductsByParentCategoryID(
	ctx context.Context,
	parentCategoryID uuid.UUID,
) ([]model.ProductSchema, error) {
	products, err := r.Queries.GetProductsByParentCategoryID(ctx, parentCategoryID)
	if err != nil {
		r.logger.Error("failed to get products by parent category ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get products by parent category ID: %w", err)
	}

	var productSchemas []model.ProductSchema
	for _, product := range products {
		productSchemas = append(productSchemas, model.ProductSchema{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description.String,
			Price:       product.Price,
			Stock:       product.Stock.Int32,
			CategoryID:  product.CategoryID.UUID,
			IsActive:    product.IsActive.Bool,
			Featured:    product.Featured.Bool,
		})
	}

	return productSchemas, nil
}

// GetProductsByFilters retrieves products by their filters
func (r *ProductRepository) GetProductsByFilters(
	ctx context.Context,
	parentCategoryUUID uuid.UUID,
	filter *utils.Filter,
	sortBy string,
) ([]database.Product, error) {
	query := `
		WITH RECURSIVE category_tree AS (
			SELECT c.id, c.name
			FROM categories c
			WHERE c.id = $1
			UNION ALL
			SELECT c.id, c.name
			FROM categories c
			INNER JOIN category_tree ct ON ct.id = c.parent_id
		)
		SELECT DISTINCT
			p.id,
			p.name,
			p.description,
			p.price,
			p.stock,
			p.category_id,
			p.created_at,
			p.updated_at,
			p.is_active,
			p.created_by,
			p.updated_by,
			p.featured
		FROM products p
		JOIN category_tree ct ON p.category_id = ct.id
		LEFT JOIN product_colors pc ON p.id = pc.product_id
	`

	args := []interface{}{parentCategoryUUID}

	if filter.HasFilter() {
		filterQuery, filterArgs := filter.GetParameterizedQuery()
		query += " WHERE " + filterQuery
		args = append(args, filterArgs...)
	}

	// Dynamic sorting
	switch sortBy {
	case "price_asc":
		query += " ORDER BY p.price ASC"
	case "price_desc":
		query += " ORDER BY p.price DESC"
	case "name_asc":
		query += " ORDER BY p.name ASC"
	case "name_desc":
		query += " ORDER BY p.name DESC"
	default:
		query += " ORDER BY p.name ASC" // Default sort
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to execute query", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var products []database.Product
	for rows.Next() {
		var product database.Product
		if err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Stock,
			&product.CategoryID,
			&product.CreatedAt,
			&product.UpdatedAt,
			&product.IsActive,
			&product.CreatedBy,
			&product.UpdatedBy,
			&product.Featured,
		); err != nil {
			r.logger.Error("failed to scan product", zap.Error(err))
			return nil, err
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("failed to iterate over rows", zap.Error(err))
		return nil, err
	}

	return products, nil
}
