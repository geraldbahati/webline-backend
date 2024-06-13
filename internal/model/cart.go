package model

import "github.com/google/uuid"

type ShoppingCart struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"userID"`
	SessionID  uuid.UUID `json:"sessionID"`
	TotalItems int32     `json:"totalItems"`
	TotalPrice float64   `json:"totalPrice"`
}

type CartItem struct {
	ID             uuid.UUID `json:"id"`
	ShoppingCartID uuid.UUID `json:"shoppingCartID"`
	ProductID      uuid.UUID `json:"productID"`
	Quantity       int32     `json:"quantity"`
	Price          string    `json:"price"`
}
