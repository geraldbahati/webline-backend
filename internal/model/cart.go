package model

import "github.com/google/uuid"

type ShoppingCart struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	SessionID  uuid.UUID
	TotalItems int32
	TotalPrice float64
}

type CartItem struct {
	ID             uuid.UUID
	ShoppingCartID uuid.UUID
	ProductID      uuid.UUID
	Quantity       int32
	Price          string
}
