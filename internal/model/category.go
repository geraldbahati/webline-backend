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

type CategoryHierarchy struct {
	Name     string              `json:"Name"`
	Children []CategoryHierarchy `json:"Children"`
}

type Category struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	ImageUrl string    `json:"imageUrl"`
}
