package utils

import (
	"weblineBackend/internal/appconfig"
	"weblineBackend/internal/model"
)

// Paginate fetches paginated data using the provided fetchData function.
func Paginate[T any](
	cfg *appconfig.Config,
	totalCount int64,
	page int32,
	pageSize int32,
	fetchData func(offset int32, limit int32) (T, error),
) (*model.PaginationResult[T], error) {

	if page < 1 {
		page = cfg.DefaultPage
	}

	// Use default for pageSize if less than 1
	if pageSize < 1 {
		pageSize = cfg.DefaultPageSize
	}

	// Adjust offset so that page 1 returns offset 0
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
