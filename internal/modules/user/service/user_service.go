// Package userservice contains the business rules for managing back-office
// users: input normalization and validation, password hashing, pagination
// limits, and translating repository errors into client-safe errors.
//
// It does not know whether it is called from GraphQL, REST or a CLI command,
// so the same rules apply everywhere.
package userservice

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/ladev707/3dbuilder-backend/internal/apperr"
	authaction "github.com/ladev707/3dbuilder-backend/internal/modules/auth/actions"
	usermodel "github.com/ladev707/3dbuilder-backend/internal/modules/user/models"
	userrepository "github.com/ladev707/3dbuilder-backend/internal/modules/user/repository"
)

// Pagination limits for List. MaxPageSize stops a client from loading the
// whole table in one request.
const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Repository is the storage the service needs.
//
// The service depends on this interface instead of the concrete Postgres
// repository. In unit tests you can pass a fake (e.g. a struct backed by a
// map) and test the business rules without a database. Go convention is to
// declare interfaces where they are used, not where they are implemented;
// *userrepository.Repository satisfies it implicitly.
type Repository interface {
	List(context.Context, usermodel.ListUsersParams) ([]usermodel.User, int, error)
	FindByID(context.Context, uuid.UUID) (usermodel.User, error)
	Create(context.Context, usermodel.User) (usermodel.User, error)
	Update(context.Context, uuid.UUID, usermodel.UserChanges) (usermodel.User, error)
	Delete(context.Context, uuid.UUID) error
}

// Service implements user CRUD use cases.
type Service struct {
	repository Repository
}

// New creates a Service backed by repository.
func New(repository Repository) *Service {
	return &Service{repository: repository}
}

// List returns one page of users. A zero limit uses DefaultPageSize; a limit
// above MaxPageSize or a negative offset is rejected with InvalidInput.
func (s *Service) List(ctx context.Context, params usermodel.ListUsersParams) (usermodel.UserPage, error) {
	if params.Limit <= 0 {
		params.Limit = DefaultPageSize
	}
	if params.Limit > MaxPageSize {
		return usermodel.UserPage{}, apperr.New(apperr.InvalidInput, "limit must not exceed 100")
	}
	if params.Offset < 0 {
		return usermodel.UserPage{}, apperr.New(apperr.InvalidInput, "offset must not be negative")
	}
	users, total, err := s.repository.List(ctx, params)
	if err != nil {
		return usermodel.UserPage{}, err
	}
	return usermodel.UserPage{Items: users, TotalCount: total}, nil
}

// Get returns one user, or an apperr NotFound error.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (usermodel.User, error) {
	user, err := s.repository.FindByID(ctx, id)
	return user, mapError(err)
}

// Create validates the input, hashes the password and stores a new user.
// The account is active unless input.IsActive is explicitly false.
func (s *Service) Create(ctx context.Context, input usermodel.CreateUserInput) (usermodel.User, error) {
	// Normalize first, then validate the normalized values, so the checks
	// apply to exactly what will be stored.
	user := usermodel.User{
		Username: authaction.NormalizeUsername(input.Username),
		Email:    authaction.NormalizeEmail(input.Email),
		// nil (not provided) defaults to true.
		IsActive: input.IsActive == nil || *input.IsActive,
	}
	// Return the first failing rule, so the client gets one clear message.
	for _, err := range []error{
		authaction.ValidateUsername(user.Username),
		authaction.ValidateEmail(user.Email),
		authaction.ValidatePassword(input.Password),
	} {
		if err != nil {
			return usermodel.User{}, err
		}
	}
	// Only the bcrypt hash is stored; the plain password goes no further.
	passwordHash, err := authaction.HashPassword(input.Password)
	if err != nil {
		return usermodel.User{}, err
	}
	user.PasswordHash = passwordHash
	user, err = s.repository.Create(ctx, user)
	return user, mapError(err)
}

// Update applies a partial update: only non-nil fields of input are
// validated and changed. An input with no fields is rejected.
func (s *Service) Update(ctx context.Context, id uuid.UUID, input usermodel.UpdateUserInput) (usermodel.User, error) {
	// Build UserChanges field by field; anything left nil is not touched.
	changes := usermodel.UserChanges{IsActive: input.IsActive}
	if input.Username != nil {
		username := authaction.NormalizeUsername(*input.Username)
		if err := authaction.ValidateUsername(username); err != nil {
			return usermodel.User{}, err
		}
		changes.Username = &username
	}
	if input.Email != nil {
		email := authaction.NormalizeEmail(*input.Email)
		if err := authaction.ValidateEmail(email); err != nil {
			return usermodel.User{}, err
		}
		changes.Email = &email
	}
	if input.Password != nil {
		if err := authaction.ValidatePassword(*input.Password); err != nil {
			return usermodel.User{}, err
		}
		passwordHash, err := authaction.HashPassword(*input.Password)
		if err != nil {
			return usermodel.User{}, err
		}
		changes.PasswordHash = &passwordHash
	}
	// Structs whose fields are all comparable can be compared with ==.
	// Equal to the zero value means every field is nil: nothing to update.
	if changes == (usermodel.UserChanges{}) {
		return usermodel.User{}, apperr.New(apperr.InvalidInput, "at least one field must be provided")
	}
	user, err := s.repository.Update(ctx, id, changes)
	return user, mapError(err)
}

// Delete soft-deletes a user and logs them out. Returns NotFound if the user
// does not exist or is already deleted.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return mapError(s.repository.Delete(ctx, id))
}

// mapError converts repository errors into client-safe apperr errors.
// Other errors (e.g. a lost database connection) pass through unchanged; the
// GraphQL error presenter logs them and shows "internal server error".
// A nil error matches no case and is returned as nil.
func mapError(err error) error {
	switch {
	case errors.Is(err, userrepository.ErrUserNotFound):
		return apperr.New(apperr.NotFound, "user not found")
	case errors.Is(err, userrepository.ErrDuplicateUser):
		return apperr.New(apperr.Conflict, "username or email already in use")
	}
	return err
}
