package model

import "github.com/google/uuid"

type ShoppingCart struct {
	ID         uuid.UUID  `json:"id"`
	UserID     *uuid.UUID `json:"userID"`
	GuestID    *uuid.UUID `json:"guestID"`
	TotalItems int32      `json:"totalItems"`
	TotalPrice float64    `json:"totalPrice"`
}

type CartItem struct {
	ID          uuid.UUID `json:"id"`
	ProductID   uuid.UUID `json:"productID"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Quantity    int32     `json:"quantity"`
	Price       string    `json:"price"`
	ImageURL    string    `json:"imageURL"`
}
