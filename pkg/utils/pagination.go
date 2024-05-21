package utils

import (
	"errors"
	"weblineBackend/internal/config"
	"weblineBackend/internal/model"
)

// Paginate fetches paginated data using the provided fetchData function.
func Paginate[T any](
	cfg *config.Config,
	totalCount int64,
	page int32,
	pageSize int32,
	fetchData func(offset int32, limit int32) (T, error),
) (*model.PaginationResult[T], error) {

	if page < 1 {
		page = cfg.DefaultPage
	}

	if pageSize < 1 {
		pageSize = cfg.DefaultPageSize
	}

	if pageSize < 1 {
		return nil, errors.New("page size must be greater than zero")
	}

	offset := (page - 1) * pageSize
	data, err := fetchData(offset, pageSize)
	if err != nil {
		return nil, err
	}

	totalPages := (totalCount + int64(pageSize) - 1) / int64(pageSize)

	return &model.PaginationResult[T]{
		TotalCount: totalCount,
		TotalPages: int32(totalPages),
		Page:       page,
		PageSize:   pageSize,
		Data:       data,
	}, nil
}
