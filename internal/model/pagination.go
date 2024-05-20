package model

type PaginationResult[T any] struct {
	TotalCount int64 `json:"total_count"`
	TotalPages int32 `json:"total_pages"`
	Page       int32 `json:"page"`
	PageSize   int32 `json:"page_size"`
	Data       T     `json:"data"`
}
