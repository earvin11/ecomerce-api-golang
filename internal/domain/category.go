package domain

import "time"

type Category struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	CreatedBy *int       `json:"created_by"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedBy *int       `json:"updated_by"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type UpdateCategory struct {
	Name *string `json:"name"`
}
