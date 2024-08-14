package sqlc

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"strconv"
	"weblineBackend/internal/database"
	"weblineBackend/internal/model"
)

type FilterProductRepoImpl struct {
	*database.Queries
	db     *sql.DB
	logger *zap.Logger
}

func NewFilterProductRepoImpl(db *sql.DB, logger *zap.Logger) *FilterProductRepoImpl {
	return &FilterProductRepoImpl{
		Queries: database.New(db),
		db:      db,
		logger:  logger,
	}
}

// execTx executes a database transaction with the provided function
func (r *FilterProductRepoImpl) execTx(ctx context.Context, fn func(*database.Queries) error) (err error) {
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

// GetTotalProductsByFilters returns the number of products that match the filter values
func (r *FilterProductRepoImpl) GetTotalProductsByFilters(ctx context.Context, filterValues *model.AllProductFilterValues) (int64, error) {
	var count int64

	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		var attributesJSON []byte

		if len(filterValues.AttributeFilters) > 0 {
			attributesJSON, err = json.Marshal(filterValues.AttributeFilters)
			if err != nil {
				r.logger.Error("failed to marshal attributes", zap.Error(err))
				return err // Properly propagate the error
			}
		}

		// Prepare parameters
		params := database.CountAllProductsByFiltersParams{
			Column1:    filterValues.CategoryNames,
			Column4:    attributesJSON,
			UsdPrice:   strconv.FormatFloat(filterValues.MinPrice, 'f', -1, 64),
			UsdPrice_2: strconv.FormatFloat(filterValues.MaxPrice, 'f', -1, 64),
		}

		// Execute the query
		count, err = q.CountAllProductsByFilters(ctx, params)
		return err
	})

	if err != nil {
		r.logger.Error("failed to get total category products by filters", zap.Error(err))
		return 0, err
	}
	return count, nil
}

// GetProductsByFilters returns products that match the filter values
func (r *FilterProductRepoImpl) GetProductsByFilters(ctx context.Context, filterValues *model.AllProductFilterValues) ([]*model.Product, error) {
	r.logger.Info("Starting GetProductsByFilters", zap.String("sortOrder", filterValues.SortOrder))

	var products []*model.Product

	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error

		switch filterValues.SortOrder {
		case "price_asc":
			products, err = r.getProductsByFiltersPriceAsc(ctx, q, filterValues)
		case "price_desc":
			products, err = r.getProductsByFiltersPriceDesc(ctx, q, filterValues)
		case "name_asc":
			products, err = r.getProductsByFiltersNameAsc(ctx, q, filterValues)
		case "name_desc":
			products, err = r.getProductsByFiltersNameDesc(ctx, q, filterValues)
		case "newest":
			products, err = r.getProductsByFiltersNewest(ctx, q, filterValues)
		case "oldest":
			products, err = r.getProductsByFiltersOldest(ctx, q, filterValues)
		default:
			products, err = r.getProductsByFiltersPriceAsc(ctx, q, filterValues)
		}
		if err != nil {
			r.logger.Error("failed to get category products by filters", zap.Error(err))
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return products, nil
}

// Individual functions for each SQLC query type

func (r *FilterProductRepoImpl) getProductsByFiltersPriceAsc(ctx context.Context, q *database.Queries, filterValues *model.AllProductFilterValues) ([]*model.Product, error) {
	var err error
	var attributesJSON []byte

	if len(filterValues.AttributeFilters) > 0 {
		attributesJSON, err = json.Marshal(filterValues.AttributeFilters)
		if err != nil {
			r.logger.Error("failed to marshal attributes", zap.Error(err))
			return nil, err
		}
	}

	rows, err := q.GetAllProductsByFiltersPriceAsc(ctx, database.GetAllProductsByFiltersPriceAscParams{
		Column1:    filterValues.CategoryNames,
		Column4:    attributesJSON,
		UsdPrice:   strconv.FormatFloat(filterValues.MinPrice, 'f', -1, 64),
		UsdPrice_2: strconv.FormatFloat(filterValues.MaxPrice, 'f', -1, 64),
		Limit:      filterValues.Limit,
		Offset:     filterValues.Offset,
	})
	if err != nil {
		return nil, err
	}

	return r.processProductRows(rows)
}

func (r *FilterProductRepoImpl) getProductsByFiltersPriceDesc(ctx context.Context, q *database.Queries, filterValues *model.AllProductFilterValues) ([]*model.Product, error) {
	var err error
	var attributesJSON []byte

	if len(filterValues.AttributeFilters) > 0 {
		attributesJSON, err = json.Marshal(filterValues.AttributeFilters)
		if err != nil {
			r.logger.Error("failed to marshal attributes", zap.Error(err))
			return nil, err
		}
	}

	rows, err := q.GetAllProductsByFiltersPriceDesc(ctx, database.GetAllProductsByFiltersPriceDescParams{
		Column1:    filterValues.CategoryNames,
		Column4:    attributesJSON,
		UsdPrice:   strconv.FormatFloat(filterValues.MinPrice, 'f', -1, 64),
		UsdPrice_2: strconv.FormatFloat(filterValues.MaxPrice, 'f', -1, 64),
		Limit:      filterValues.Limit,
		Offset:     filterValues.Offset,
	})
	if err != nil {
		return nil, err
	}

	return r.processProductRows(rows)
}

func (r *FilterProductRepoImpl) getProductsByFiltersNameAsc(ctx context.Context, q *database.Queries, filterValues *model.AllProductFilterValues) ([]*model.Product, error) {
	var err error
	var attributesJSON []byte

	if len(filterValues.AttributeFilters) > 0 {
		attributesJSON, err = json.Marshal(filterValues.AttributeFilters)
		if err != nil {
			r.logger.Error("failed to marshal attributes", zap.Error(err))
			return nil, err
		}
	}

	rows, err := q.GetAllProductsByFiltersNameAsc(ctx, database.GetAllProductsByFiltersNameAscParams{
		Column1:    filterValues.CategoryNames,
		Column4:    attributesJSON,
		UsdPrice:   strconv.FormatFloat(filterValues.MinPrice, 'f', -1, 64),
		UsdPrice_2: strconv.FormatFloat(filterValues.MaxPrice, 'f', -1, 64),
		Limit:      filterValues.Limit,
		Offset:     filterValues.Offset,
	})
	if err != nil {
		return nil, err
	}

	return r.processProductRows(rows)
}

func (r *FilterProductRepoImpl) getProductsByFiltersNameDesc(ctx context.Context, q *database.Queries, filterValues *model.AllProductFilterValues) ([]*model.Product, error) {
	var err error
	var attributesJSON []byte

	if len(filterValues.AttributeFilters) > 0 {
		attributesJSON, err = json.Marshal(filterValues.AttributeFilters)
		if err != nil {
			r.logger.Error("failed to marshal attributes", zap.Error(err))
			return nil, err
		}
	}

	rows, err := q.GetAllProductsByFiltersNameDesc(ctx, database.GetAllProductsByFiltersNameDescParams{
		Column1:    filterValues.CategoryNames,
		Column4:    attributesJSON,
		UsdPrice:   strconv.FormatFloat(filterValues.MinPrice, 'f', -1, 64),
		UsdPrice_2: strconv.FormatFloat(filterValues.MaxPrice, 'f', -1, 64),
		Limit:      filterValues.Limit,
		Offset:     filterValues.Offset,
	})
	if err != nil {
		return nil, err
	}

	return r.processProductRows(rows)
}

func (r *FilterProductRepoImpl) getProductsByFiltersNewest(ctx context.Context, q *database.Queries, filterValues *model.AllProductFilterValues) ([]*model.Product, error) {
	var err error
	var attributesJSON []byte

	if len(filterValues.AttributeFilters) > 0 {
		attributesJSON, err = json.Marshal(filterValues.AttributeFilters)
		if err != nil {
			r.logger.Error("failed to marshal attributes", zap.Error(err))
			return nil, err
		}
	}

	rows, err := q.GetAllProductsByFiltersNewest(ctx, database.GetAllProductsByFiltersNewestParams{
		Column1:    filterValues.CategoryNames,
		Column4:    attributesJSON,
		UsdPrice:   strconv.FormatFloat(filterValues.MinPrice, 'f', -1, 64),
		UsdPrice_2: strconv.FormatFloat(filterValues.MaxPrice, 'f', -1, 64),
		Limit:      filterValues.Limit,
		Offset:     filterValues.Offset,
	})
	if err != nil {
		return nil, err
	}

	return r.processProductRows(rows)
}

func (r *FilterProductRepoImpl) getProductsByFiltersOldest(ctx context.Context, q *database.Queries, filterValues *model.AllProductFilterValues) ([]*model.Product, error) {
	var err error
	var attributesJSON []byte

	if len(filterValues.AttributeFilters) > 0 {
		attributesJSON, err = json.Marshal(filterValues.AttributeFilters)
		if err != nil {
			r.logger.Error("failed to marshal attributes", zap.Error(err))
			return nil, err
		}
	}

	rows, err := q.GetAllProductsByFiltersOldest(ctx, database.GetAllProductsByFiltersOldestParams{
		Column1:    filterValues.CategoryNames,
		Column4:    attributesJSON,
		UsdPrice:   strconv.FormatFloat(filterValues.MinPrice, 'f', -1, 64),
		UsdPrice_2: strconv.FormatFloat(filterValues.MaxPrice, 'f', -1, 64),
		Limit:      filterValues.Limit,
		Offset:     filterValues.Offset,
	})
	if err != nil {
		return nil, err
	}

	return r.processProductRows(rows)
}

// processProductRows processes each row of the product result and converts it to a model.Product
func (r *FilterProductRepoImpl) processProductRows(rows interface{}) ([]*model.Product, error) {
	var products []*model.Product

	switch rows := rows.(type) {
	case []database.GetAllProductsByFiltersPriceAscRow:
		for _, row := range rows {
			products = append(products, r.convertToProduct(row.ID, row.Name, row.Description.String, row.Price, row.Imageurl, row.Discountpercent, row.Slug))
		}
	case []database.GetAllProductsByFiltersPriceDescRow:
		for _, row := range rows {
			products = append(products, r.convertToProduct(row.ID, row.Name, row.Description.String, row.Price, row.Imageurl, row.Discountpercent, row.Slug))
		}
	case []database.GetAllProductsByFiltersNameAscRow:
		for _, row := range rows {
			products = append(products, r.convertToProduct(row.ID, row.Name, row.Description.String, row.Price, row.Imageurl, row.Discountpercent, row.Slug))
		}
	case []database.GetAllProductsByFiltersNameDescRow:
		for _, row := range rows {
			products = append(products, r.convertToProduct(row.ID, row.Name, row.Description.String, row.Price, row.Imageurl, row.Discountpercent, row.Slug))
		}
	case []database.GetAllProductsByFiltersNewestRow:
		for _, row := range rows {
			products = append(products, r.convertToProduct(row.ID, row.Name, row.Description.String, row.Price, row.Imageurl, row.Discountpercent, row.Slug))
		}
	case []database.GetAllProductsByFiltersOldestRow:
		for _, row := range rows {
			products = append(products, r.convertToProduct(row.ID, row.Name, row.Description.String, row.Price, row.Imageurl, row.Discountpercent, row.Slug))
		}
	}

	return products, nil
}

// convertToProduct converts a row to a model.Product
func (r *FilterProductRepoImpl) convertToProduct(id uuid.UUID, name, description string, price string, imageURL, discountPercent, slug string) *model.Product {
	discount, err := strconv.ParseFloat(discountPercent, 64)
	if err != nil {
		r.logger.Error("failed to parse discount percent", zap.Error(err))
		return nil
	}

	return &model.Product{
		ID:              id,
		Name:            name,
		Description:     description,
		Price:           price,
		ImageURL:        imageURL,
		DiscountPercent: discount,
		Slug:            slug,
	}
}

// GetProductAttributes returns the attributes and total products for a category
func (r *FilterProductRepoImpl) GetProductAttributes(ctx context.Context) (*model.FilterOptions, error) {
	var filterOptions model.FilterOptions

	err := r.execTx(ctx, func(q *database.Queries) error {
		var err error
		filterOptions, err = r.getProductAttributes(ctx, q)
		return err
	})

	if err != nil {
		return nil, err
	}

	return &filterOptions, nil
}

// getProductAttributesAndCountByCategoryID returns the attributes and total products for a category
func (r *FilterProductRepoImpl) getProductAttributes(ctx context.Context, q *database.Queries) (model.FilterOptions, error) {
	rows, err := q.GetProductAttributes(ctx)
	if err != nil {
		return model.FilterOptions{}, err
	}

	return r.processFilterOptionsRow(rows), nil
}

func (r *FilterProductRepoImpl) processFilterOptionsRow(row database.GetProductAttributesRow) model.FilterOptions {
	attributes := make(map[string][]string)

	// Unmarshal the JSONB attributes to a map[string][]string
	if err := json.Unmarshal(row.Attributes, &attributes); err != nil {
		r.logger.Error("failed to unmarshal attributes", zap.Error(err))
		return model.FilterOptions{
			Attributes:    attributes,
			TotalProducts: row.TotalProducts,
		}
	}

	return model.FilterOptions{
		Attributes:    attributes,
		TotalProducts: row.TotalProducts,
	}
}
