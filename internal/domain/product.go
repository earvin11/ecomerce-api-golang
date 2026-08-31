package domain

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

type Cents int64

func (c Cents) MarshalJSON() ([]byte, error) {
	sign := ""
	v := int64(c)
	if v < 0 {
		sign = "-"
		v = -v
	}
	return []byte(fmt.Sprintf("%s%d.%02d", sign, v/100, v%100)), nil
}

func (c *Cents) UnmarshalJSON(data []byte) error {
	f, err := strconv.ParseFloat(string(data), 64)
	if err != nil {
		return fmt.Errorf("price must be a decimal number: %w", err)
	}
	*c = Cents(math.Round(f * 100))
	return nil
}

type Product struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	Price     Cents      `json:"price"`
	Category  int        `json:"category"`
	InStock   bool       `json:"in_stock"`
	Quantity  int        `json:"quantity"`
	Img       *string    `json:"img"`
	CreatedBy *int       `json:"created_by"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedBy *int       `json:"updated_by"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type UpdateProduct struct {
	Name     *string        `json:"name"`
	Price    *Cents         `json:"price"`
	Category *int           `json:"category"`
	InStock  *bool          `json:"in_stock"`
	Quantity *int           `json:"quantity"`
	Img      NullableString `json:"img"`
}
