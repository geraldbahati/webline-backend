package model

import (
	"time"

	"github.com/google/uuid"
)

type CreateGuestCheckoutParams struct {
	Email         string `json:"email"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Phone         string `json:"phone"`
	StreetAddress string `json:"street_address"`
	City          string `json:"city"`
	State         string `json:"state"`
	Country       string `json:"country"`
}

type CreateOrderItemParams struct {
	ProductID        uuid.UUID       `json:"product_id"`
	ProductOptionIDs []uuid.NullUUID `json:"product_option_id"`
	ColorID          uuid.NullUUID   `json:"color_id"`
	SizeID           uuid.NullUUID   `json:"size_id"`
	Quantity         int32           `json:"quantity"`
	Price            string          `json:"price"`
}

var AVAILABLE_COUNTRIES = []string{"Kenya"}

var COUNTIES = []string{
	"Mombasa", "Kwale", "Kilifi", "Tana River", "Lamu", "Taita-Taveta", "Garissa",
	"Wajir", "Mandera", "Marsabit", "Isiolo", "Meru", "Tharaka-Nithi", "Embu",
	"Kitui", "Machakos", "Makueni", "Nyandarua", "Nyeri", "Kirinyaga", "Murang'a",
	"Kiambu", "Turkana", "West Pokot", "Samburu", "Trans-Nzoia", "Uasin Gishu",
	"Elgeyo-Marakwet", "Nandi", "Baringo", "Laikipia", "Nakuru", "Narok", "Kajiado",
	"Kericho", "Bomet", "Kakamega", "Vihiga", "Bungoma", "Busia", "Siaya", "Kisumu",
	"Homa Bay", "Migori", "Kisii", "Nyamira", "Nairobi",
}

const (
	MIN_NAME_LENGTH     = 2
	MAX_NAME_LENGTH     = 50
	MIN_PHONE_LENGTH    = 10
	MAX_PHONE_LENGTH    = 15
	MIN_PASSWORD_LENGTH = 6
)

type CreateOrderParams struct {
	GuestID          *uuid.UUID `json:"guest_id"`
	UserID           *uuid.UUID `json:"user_id"`
	FirstName        string     `json:"first_name" validate:"required,min=2,max=50"`
	LastName         string     `json:"last_name" validate:"required,min=2,max=50"`
	Country          string     `json:"country" validate:"required,oneof=Kenya"`
	KraPIN           *string    `json:"kraPIN,omitempty"`
	CompanyName      *string    `json:"companyName,omitempty"`
	City             string     `json:"city" validate:"required,min=2,max=50"`
	County           string     `json:"county" validate:"required,oneof=Mombasa Kwale Kilifi 'Tana River' Lamu Taita-Taveta Garissa Wajir Mandera Marsabit Isiolo Meru Tharaka-Nithi Embu Kitui Machakos Makueni Nyandarua Nyeri Kirinyaga Murang'a Kiambu Turkana 'West Pokot' Samburu Trans-Nzoia 'Uasin Gishu' Elgeyo-Marakwet Nandi Baringo Laikipia Nakuru Narok Kajiado Kericho Bomet Kakamega Vihiga Bungoma Busia Siaya Kisumu 'Homa Bay' Migori Kisii Nyamira Nairobi"`
	Phone            string     `json:"phone" validate:"required,min=10,max=15,numeric"`
	Email            string     `json:"email" validate:"required,email"`
	CanCreateAccount bool       `json:"canCreateAccount"`
	Password         *string    `json:"password,omitempty" validate:"omitempty,min=6"`
	OrderNotes       *string    `json:"orderNotes,omitempty"`
	GrandTotal       float64    `json:"grandTotal"`
	DiscountAmount   float64    `json:"discountAmount"`
	VatAmount        float64    `json:"vatAmount"`
	SubTotal         float64    `json:"subTotal"`
	OrderDate        time.Time  `json:"orderDate"`
	OrderNumber      string     `json:"orderNumber"`
}

type OrderClientResponse struct {
	ID             uuid.UUID
	OrderNumber    string
	OrderCreatedAt time.Time
	CustomerName   string
	Phone          string
	Amount         float64
}

type Revenue struct {
	Revenue       float64 `json:"revenue"`
	MonthlyGrowth float64 `json:"monthlyGrowth"`
}

type MonthlyRevenue struct {
	Month   time.Time `json:"month"`
	Revenue float64   `json:"revenue"`
}

type OrderUser struct {
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Amount   float64 `json:"amount"`
	Fallback string  `json:"fallback"`
}

type OrderAmounts struct {
	SubTotal     float64 `json:"subTotal"`
	TaxAmount    float64 `json:"taxAmount"`
	ShippingAmount float64 `json:"shippingAmount"`
	DiscountAmount float64 `json:"discountAmount"`
	VatAmount    float64 `json:"vatAmount"`
	GrandTotal   float64 `json:"grandTotal"`
}

type OrderSchema struct {
	ID             uuid.UUID
	OrderNumber    string
	CreatedAt      time.Time
}
