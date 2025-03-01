package model

import (
	"github.com/google/uuid"
)

type CategoryDetail struct {
	ID              uuid.UUID        `json:"id"`
	Name            string           `json:"name"`
	ParentID        uuid.UUID        `json:"parentID"`
	IsActive        bool             `json:"isActive"`
	ImageURL        string           `json:"imageURL"`
	SubCategories   []CategoryDetail `json:"subCategories"`
	AvailableColors []string         `json:"availableColors"`
}

type V2CategoryDetail struct {
	Slug            string `json:"slug"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	ParentID        string `json:"categoryID"`
	MetaTitle       string `json:"metaTitle"`
	MetaDescription string `json:"metaDescription"`
	ImageURL        string `json:"imageURL"`
}

type CategorySEO struct {
	MetaTitle       string `json:"metaTitle"`
	MetaDescription string `json:"metaDescription"`
}

type CategoryHierarchy struct {
	Name     string              `json:"Name"`
	Children []CategoryHierarchy `json:"Children"`
}

type V2CategoryHierarchy struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Position         int                    `json:"position"`
	NumberOfProducts int                    `json:"numberOfProducts"`
	IsActive         bool                   `json:"isActive"`
	Slug             string                 `json:"slug"`
	Children         []*V2CategoryHierarchy `json:"children,omitempty"`
}

type Category struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	ImageUrl string    `json:"imageUrl"`
}

type Collection struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	ParentName string    `json:"parentName"`
	ImageUrl   string    `json:"imageUrl"`
}

type ColorMetafield struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Value string    `json:"value"`
}

type ProcessorMetafield struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type StorageMetafield struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type SizeMetafield struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type ProductMetafields struct {
	Color     []ColorMetafield     `json:"color"`
	Processor []ProcessorMetafield `json:"processor"`
	Size      []SizeMetafield      `json:"size"`
	Storage   []StorageMetafield   `json:"storage"`
}

type CreateCategoryParams struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	MetaTitle       string `json:"metaTitle"`
	MetaDescription string `json:"metaDescription"`
	ParentID        string `json:"parentID"`
	Slug            string `json:"slug"`
}

type CategoryStats struct {
	DirectChildrenCount int `json:"directChildrenCount"`
	AllDescendantsCount int `json:"allDescendantsCount"`
	DirectProductCount  int `json:"directProductCount"`
	TotalProductCount   int `json:"totalProductCount"`
	DepthLevel          int `json:"depthLevel"`
}

type CategoryWithProductCount struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	ImageURL     string `json:"imageUrl"`
	Description  string `json:"description"`
	IsActive     bool   `json:"isActive"`
	Position     int    `json:"position"`
	Depth        int    `json:"depth"`
	ProductCount int    `json:"productCount"`
}

type CategoryHierarchyStats struct {
	Category CategoryDetail             `json:"category"`
	Stats    CategoryStats              `json:"stats"`
	Children []CategoryWithProductCount `json:"children"`
}

// CategoryBreadcrumb represents one level in the category ancestry path
type CategoryBreadcrumb struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// CategoryWithAncestry contains category information with its full ancestry path
type CategoryWithAncestry struct {
	ID           string               `json:"id"`
	Name         string               `json:"name"`
	Slug         string               `json:"slug"`
	Description  string               `json:"description"`
	IsActive     bool                 `json:"isActive"`
	ProductCount int                  `json:"productCount"`
	AncestryPath []CategoryBreadcrumb `json:"ancestryPath"`
}
