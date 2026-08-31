package domain

import "time"

type Purchase struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	ProductID   int       `json:"product_id"`
	ProductName string    `json:"product_name,omitempty"`
	Quantity    int       `json:"quantity"`
	UnitPrice   Cents     `json:"unit_price"`
	Discount    Cents     `json:"discount"`
	Total       Cents     `json:"total"`
	WalletRef   *string   `json:"wallet_ref"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreatePurchase struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}
