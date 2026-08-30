package domain

type User struct {
	ID       int     `json:"id"`
	Email    string  `json:"email"`
	Password string  `json:"-"`
	Username string  `json:"username"`
	RoleID   int     `json:"role_id"`
	Role     *Role   `json:"role,omitempty"`
	Photo    *string `json:"photo"`
}

type CreateUser struct {
	Email    string  `json:"email"`
	Password string  `json:"password"`
	Username string  `json:"username"`
	RoleID   int     `json:"role_id"`
	Photo    *string `json:"photo"`
}

type UpdateUser struct {
	Email    *string        `json:"email"`
	Password *string        `json:"password"`
	Username *string        `json:"username"`
	RoleID   *int           `json:"role_id"`
	Photo    NullableString `json:"photo"`
}
