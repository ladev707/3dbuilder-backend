// Package customermodel contains the plain data types of the customer module.
//
// It mirrors usermodel. gqlgen binds the GraphQL types Customer,
// CustomerPage, CustomerFilter, CreateCustomerInput and UpdateCustomerInput
// to these structs by name through "autobind" in gqlgen.yml.
package customermodel

import (
	"time"

	"github.com/google/uuid"
)

// Customer is a storefront account (a row in the customers table). Customers
// log in through the "customer" auth gate and cannot use the admin GraphQL API.
type Customer struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	// PasswordHash is the bcrypt hash and is never serialized.
	PasswordHash string     `json:"-"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

// ClientRole links a customer to a role name.
type ClientRole struct {
	ID   uuid.UUID `json:"id"`
	Role string    `json:"role"`
}

// CreateCustomerInput is the data a client sends to create a customer.
type CreateCustomerInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	// IsActive defaults to true when nil (not provided).
	IsActive *bool `json:"is_active"`
}

// UpdateCustomerInput is the data a client sends to update a customer.
// Nil fields are left unchanged, which allows partial updates.
type UpdateCustomerInput struct {
	Username *string `json:"username"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
	IsActive *bool   `json:"is_active"`
}

// CustomerChanges is the validated update the service hands to the
// repository. It carries the hashed password, never the plain one.
type CustomerChanges struct {
	Username     *string
	Email        *string
	PasswordHash *string
	IsActive     *bool
}

// CustomerFilter narrows a customer list. Nil fields mean "do not filter".
type CustomerFilter struct {
	// Search matches username or email, case-insensitively.
	Search   *string `json:"search"`
	IsActive *bool   `json:"is_active"`
}

// ListCustomersParams combines the filter with offset pagination.
type ListCustomersParams struct {
	Filter CustomerFilter
	Limit  int
	Offset int
}

// CustomerPage is one page of customers plus the total number of matches.
type CustomerPage struct {
	Items      []Customer `json:"items"`
	TotalCount int        `json:"total_count"`
}
