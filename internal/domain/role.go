package domain

type RoleNames map[string]bool

const (
	RoleAdmin    = "ADMIN"
	RoleEditor   = "EDITOR"
	RoleCustomer = "CUSTOMER"
)

func (r RoleNames) Has(name string) bool {
	return r[name]
}

var AllowedRoles = RoleNames{
	RoleAdmin:    true,
	RoleCustomer: true,
	RoleEditor:   true,
}

func IsDiscountedRole(role string) bool {
	return role == RoleAdmin || role == RoleEditor
}

type Role struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type UpdateRole struct {
	Name *string `json:"name"`
}
