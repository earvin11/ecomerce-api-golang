package domain

type RoleNames map[string]bool

var AllowedRoles = RoleNames{
	"ADMIN":    true,
	"CUSTOMER": true,
	"EDITOR":   true,
}

type Role struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type UpdateRole struct {
	Name *string `json:"name"`
}
