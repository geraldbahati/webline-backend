package repository

import (
	"context"
	"database/sql"
	"fmt"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"

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
func (r *ProductRepository) ListProducts(ctx context.Context, limit int32, offset int32, sortBy string, categories []string, priceFrom, priceTo string) ([]model.ProductSchema, error) {
	var products []model.ProductSchema
	switch sortBy {
	case "price_asc":
		filterProducts, err := r.Queries.GetFilteredProductsOrderByPriceAsc(ctx, database.GetFilteredProductsOrderByPriceAscParams{
			Column1: categories,
			Price:   priceFrom,
			Price_2: priceTo,
			Limit:   limit,
			Offset:  offset,
		})
		if err != nil {
			r.logger.Error("failed to list products", zap.Error(err))
			return nil, fmt.Errorf("failed to list products: %w", err)
		}

		for _, product := range filterProducts {
			products = append(products, model.ProductSchema{
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

	case "price_desc":
		filterProducts, err := r.Queries.GetFilteredProductsOrderByPriceDesc(ctx, database.GetFilteredProductsOrderByPriceDescParams{
			Column1: categories,
			Price:   priceFrom,
			Price_2: priceTo,
			Limit:   limit,
			Offset:  offset,
		})
		if err != nil {
			r.logger.Error("failed to list products", zap.Error(err))
			return nil, fmt.Errorf("failed to list products: %w", err)
		}

		for _, product := range filterProducts {
			products = append(products, model.ProductSchema{
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

	case "name_asc":
		filterProducts, err := r.Queries.GetFilteredProductsOrderByNameAsc(ctx, database.GetFilteredProductsOrderByNameAscParams{
			Column1: categories,
			Price:   priceFrom,
			Price_2: priceTo,
			Limit:   limit,
			Offset:  offset,
		})
		if err != nil {
			r.logger.Error("failed to list products", zap.Error(err))
			return nil, fmt.Errorf("failed to list products: %w", err)
		}

		for _, product := range filterProducts {
			products = append(products, model.ProductSchema{
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

	case "name_desc":
		filterProducts, err := r.Queries.GetFilteredProductsOrderByNameDesc(ctx, database.GetFilteredProductsOrderByNameDescParams{
			Column1: categories,
			Price:   priceFrom,
			Price_2: priceTo,
			Limit:   limit,
			Offset:  offset,
		})
		if err != nil {
			r.logger.Error("failed to list products", zap.Error(err))
			return nil, fmt.Errorf("failed to list products: %w", err)
		}

		for _, product := range filterProducts {
			products = append(products, model.ProductSchema{
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

	default:
		filterProducts, err := r.Queries.GetFilteredProducts(ctx, database.GetFilteredProductsParams{
			Column1: categories,
			Price:   priceFrom,
			Price_2: priceTo,
			Limit:   limit,
			Offset:  offset,
		})
		if err != nil {
			r.logger.Error("failed to list products", zap.Error(err))
			return nil, fmt.Errorf("failed to list products: %w", err)
		}

		for _, product := range filterProducts {
			products = append(products, model.ProductSchema{
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
	}

	return products, nil
}

// CountFilteredProducts returns the total number of products that match the filter
func (r *ProductRepository) CountFilteredProducts(
	ctx context.Context,
	categories []string,
	priceFrom, priceTo string,
) (int64, error) {
	count, err := r.Queries.CountFilteredProducts(ctx, database.CountFilteredProductsParams{
		Column1: categories,
		Price:   priceFrom,
		Price_2: priceTo,
	})
	if err != nil {
		r.logger.Error("failed to count products", zap.Error(err))
		return 0, fmt.Errorf("failed to count products: %w", err)
	}

	return count, nil
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
	limit int32,
	offset int32,
) ([]model.ProductSchema, error) {
	products, err := r.Queries.GetProductsByParentCategoryID(ctx, database.GetProductsByParentCategoryIDParams{
		ID:     categoryID,
		Limit:  limit,
		Offset: offset,
	})
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

// CountProductsByParentCategoryID counts the number of products in a category
func (r *ProductRepository) CountProductsByParentCategoryID(
	ctx context.Context,
	categoryID uuid.UUID,
) (int64, error) {
	count, err := r.Queries.CountProductsByParentCategoryID(ctx, categoryID)
	if err != nil {
		r.logger.Error("failed to count products by category ID", zap.Error(err))
		return 0, fmt.Errorf("failed to count products by category ID: %w", err)
	}

	return count, nil
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
	products, err := r.Queries.GetProductsByParentCategoryID(ctx, database.GetProductsByParentCategoryIDParams{
		ID:     parentCategoryID,
		Limit:  0,
		Offset: 100,
	})
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
func (r *ProductRepository) GetProductsByFiltersPriceAsc(ctx context.Context, params database.GetProductsByFiltersPriceAscParams) ([]database.GetProductsByFiltersPriceAscRow, error) {
	return r.Queries.GetProductsByFiltersPriceAsc(ctx, params)
}

func (r *ProductRepository) GetProductsByFiltersPriceDesc(ctx context.Context, params database.GetProductsByFiltersPriceDescParams) ([]database.GetProductsByFiltersPriceDescRow, error) {
	return r.Queries.GetProductsByFiltersPriceDesc(ctx, params)
}

func (r *ProductRepository) GetProductsByFiltersNameAsc(ctx context.Context, params database.GetProductsByFiltersNameAscParams) ([]database.GetProductsByFiltersNameAscRow, error) {
	return r.Queries.GetProductsByFiltersNameAsc(ctx, params)
}

func (r *ProductRepository) GetProductsByFiltersNameDesc(ctx context.Context, params database.GetProductsByFiltersNameDescParams) ([]database.GetProductsByFiltersNameDescRow, error) {
	return r.Queries.GetProductsByFiltersNameDesc(ctx, params)
}

func (r *ProductRepository) GetProductsByFiltersDefault(ctx context.Context, params database.GetProductsByFiltersDefaultParams) ([]database.GetProductsByFiltersDefaultRow, error) {
	return r.Queries.GetProductsByFiltersDefault(ctx, params)
}

func (r *ProductRepository) GetProductsByFiltersNewest(ctx context.Context, params database.GetProductsByFiltersNewestParams) ([]database.GetProductsByFiltersNewestRow, error) {
	return r.Queries.GetProductsByFiltersNewest(ctx, params)
}

func (r *ProductRepository) GetProductsByFiltersOldest(ctx context.Context, params database.GetProductsByFiltersOldestParams) ([]database.GetProductsByFiltersOldestRow, error) {
	return r.Queries.GetProductsByFiltersOldest(ctx, params)
}

func (r *ProductRepository) GetFilterOptions(ctx context.Context) ([]database.GetFilterOptionsRow, error) {
	return r.Queries.GetFilterOptions(ctx)
}
