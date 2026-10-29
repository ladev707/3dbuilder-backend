// Package customerservice contains the business rules for managing
// customers. It mirrors userservice; the one difference is that new
// customers are always given DefaultRole.
package customerservice

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/ladev707/3dbuilder-backend/internal/apperr"
	authaction "github.com/ladev707/3dbuilder-backend/internal/modules/auth/actions"
	customermodel "github.com/ladev707/3dbuilder-backend/internal/modules/customer/models"
	customerrepository "github.com/ladev707/3dbuilder-backend/internal/modules/customer/repository"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
	// DefaultRole is assigned to every new customer. It is seeded in
	// migrations/00001_auth_rbac.sql with the profile.read and profile.update
	// permissions.
	DefaultRole = "customer"
)

// Repository is the storage the service needs. Declaring it here lets tests
// pass a fake implementation instead of Postgres.
type Repository interface {
	List(context.Context, customermodel.ListCustomersParams) ([]customermodel.Customer, int, error)
	FindByID(context.Context, uuid.UUID) (customermodel.Customer, error)
	Create(context.Context, customermodel.Customer, string) (customermodel.Customer, error)
	Update(context.Context, uuid.UUID, customermodel.CustomerChanges) (customermodel.Customer, error)
	Delete(context.Context, uuid.UUID) error
}

// Service implements customer CRUD use cases.
type Service struct {
	repository Repository
}

// New creates a Service backed by repository.
func New(repository Repository) *Service {
	return &Service{repository: repository}
}

// List returns one page of customers, applying the default and maximum page size.
func (s *Service) List(ctx context.Context, params customermodel.ListCustomersParams) (customermodel.CustomerPage, error) {
	if params.Limit <= 0 {
		params.Limit = DefaultPageSize
	}
	if params.Limit > MaxPageSize {
		return customermodel.CustomerPage{}, apperr.New(apperr.InvalidInput, "limit must not exceed 100")
	}
	if params.Offset < 0 {
		return customermodel.CustomerPage{}, apperr.New(apperr.InvalidInput, "offset must not be negative")
	}
	customers, total, err := s.repository.List(ctx, params)
	if err != nil {
		return customermodel.CustomerPage{}, err
	}
	return customermodel.CustomerPage{Items: customers, TotalCount: total}, nil
}

// Get returns one customer, or an apperr NotFound error.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (customermodel.Customer, error) {
	customer, err := s.repository.FindByID(ctx, id)
	return customer, mapError(err)
}

// Create validates the input, hashes the password and stores a new customer
// with DefaultRole, so they can log in with the standard permissions.
func (s *Service) Create(ctx context.Context, input customermodel.CreateCustomerInput) (customermodel.Customer, error) {
	customer := customermodel.Customer{
		Username: authaction.NormalizeUsername(input.Username),
		Email:    authaction.NormalizeEmail(input.Email),
		IsActive: input.IsActive == nil || *input.IsActive,
	}
	for _, err := range []error{
		authaction.ValidateUsername(customer.Username),
		authaction.ValidateEmail(customer.Email),
		authaction.ValidatePassword(input.Password),
	} {
		if err != nil {
			return customermodel.Customer{}, err
		}
	}
	passwordHash, err := authaction.HashPassword(input.Password)
	if err != nil {
		return customermodel.Customer{}, err
	}
	customer.PasswordHash = passwordHash
	customer, err = s.repository.Create(ctx, customer, DefaultRole)
	return customer, mapError(err)
}

// Update applies a partial update: only non-nil fields are validated and
// changed. An input with no fields is rejected.
func (s *Service) Update(ctx context.Context, id uuid.UUID, input customermodel.UpdateCustomerInput) (customermodel.Customer, error) {
	changes := customermodel.CustomerChanges{IsActive: input.IsActive}
	if input.Username != nil {
		username := authaction.NormalizeUsername(*input.Username)
		if err := authaction.ValidateUsername(username); err != nil {
			return customermodel.Customer{}, err
		}
		changes.Username = &username
	}
	if input.Email != nil {
		email := authaction.NormalizeEmail(*input.Email)
		if err := authaction.ValidateEmail(email); err != nil {
			return customermodel.Customer{}, err
		}
		changes.Email = &email
	}
	if input.Password != nil {
		if err := authaction.ValidatePassword(*input.Password); err != nil {
			return customermodel.Customer{}, err
		}
		passwordHash, err := authaction.HashPassword(*input.Password)
		if err != nil {
			return customermodel.Customer{}, err
		}
		changes.PasswordHash = &passwordHash
	}
	// All fields nil means the client sent an empty input.
	if changes == (customermodel.CustomerChanges{}) {
		return customermodel.Customer{}, apperr.New(apperr.InvalidInput, "at least one field must be provided")
	}
	customer, err := s.repository.Update(ctx, id, changes)
	return customer, mapError(err)
}

// Delete soft-deletes a customer and logs them out.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return mapError(s.repository.Delete(ctx, id))
}

// mapError converts repository errors into client-safe apperr errors.
// Anything else, including ErrRoleNotFound (a setup problem, not the
// client's fault), passes through and is reported as an internal error.
func mapError(err error) error {
	switch {
	case errors.Is(err, customerrepository.ErrCustomerNotFound):
		return apperr.New(apperr.NotFound, "customer not found")
	case errors.Is(err, customerrepository.ErrDuplicateCustomer):
		return apperr.New(apperr.Conflict, "username or email already in use")
	}
	return err
}
