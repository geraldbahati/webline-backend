package model

import "github.com/google/uuid"

type CategoryProductFilterValues struct {
	CategoryID       uuid.UUID
	CategoryNames    []string
	AttributeFilters map[string][]string
	MinPrice         float64
	MaxPrice         float64
	SortOrder        string
	Limit            int32
	Offset           int32
}

type AllProductFilterValues struct {
	CategoryNames    []string
	AttributeFilters map[string][]string
	MinPrice         float64
	MaxPrice         float64
	SortOrder        string
	Limit            int32
	Offset           int32
}

type FilterOptions struct {
	Attributes    map[string][]string `json:"attributes"`
	TotalProducts int64               `json:"totalProducts"`
}
