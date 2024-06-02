package model

import (
	"github.com/google/uuid"
)

type Category struct {
	ID              uuid.UUID  `json:"ID"`
	Name            string     `json:"Name"`
	ParentID        uuid.UUID  `json:"ParentID"`
	IsActive        bool       `json:"IsActive"`
	SubCategories   []Category `json:"SubCategories"`
	AvailableColors []string   `json:"AvailableColors"`
}

type CategoryHierarchy struct {
	Name     string              `json:"Name"`
	Children []CategoryHierarchy `json:"Children"`
}
