package model

type PaginationResult[T any] struct {
	TotalCount int64 `json:"totalCount"`
	TotalPages int32 `json:"totalPages"`
	Page       int32 `json:"page"`
	PageSize   int32 `json:"pageSize"`
	Data       T     `json:"data"`
}
