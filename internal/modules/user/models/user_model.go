// These structs are shared by every layer: the repository fills them from
// SQL, the service applies business rules to them, and gqlgen binds the
// GraphQL types with the same names (User, UserPage, UserFilter,
// CreateUserInput, UpdateUserInput) straight to them through "autobind" in
// gqlgen.yml. Renaming a struct or field here therefore changes the GraphQL
// mapping, and `make graphql` will tell you if something no longer matches.
package usermodel

import (
	"time"

	"github.com/google/uuid"
)

// User is a back-office account (a row in the users table).
type User struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	// PasswordHash is the bcrypt hash. `json:"-"` stops it from ever being
	// serialized, and it is not part of the GraphQL schema either.
	PasswordHash string     `json:"-"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

// UserRole links a user to a role name.
type UserRole struct {
	ID   uuid.UUID `json:"id"`
	Role string    `json:"role"`
}

// CreateUserInput is the data a client sends to create a user
// (GraphQL input CreateUserInput).
type CreateUserInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	// IsActive is a pointer so "not provided" (nil) can be told apart from
	// false. The service treats nil as true.
	IsActive *bool `json:"is_active"`
}

// UpdateUserInput is the data a client sends to update a user
// (GraphQL input UpdateUserInput). Every field is a pointer: nil means
// "leave this field unchanged", which allows partial updates.
type UpdateUserInput struct {
	Username *string `json:"username"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
	IsActive *bool   `json:"is_active"`
}

// UserChanges is what the service hands to the repository after validating
// an UpdateUserInput. It differs from the input in one important way: it
// carries the hashed password, never the plain one, so the repository cannot
// store a plain password by mistake. Nil fields are left unchanged.
type UserChanges struct {
	Username     *string
	Email        *string
	PasswordHash *string
	IsActive     *bool
}

// UserFilter narrows a user list (GraphQL input UserFilter).
// Nil fields mean "do not filter on this".
type UserFilter struct {
	// Search matches username or email, case-insensitively.
	Search   *string `json:"search"`
	IsActive *bool   `json:"is_active"`
}

// ListUsersParams combines the filter with offset pagination.
type ListUsersParams struct {
	Filter UserFilter
	// Limit is the page size. The service applies the default and maximum.
	Limit int
	// Offset is how many matching rows to skip.
	Offset int
}

// UserPage is one page of users plus the total number of matches, which
// clients need to render pagination (GraphQL type UserPage).
type UserPage struct {
	Items      []User `json:"items"`
	TotalCount int    `json:"total_count"`
}
