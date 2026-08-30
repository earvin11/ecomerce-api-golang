package domain

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type UpdateCategory struct {
	Name *string `json:"name"`
}
