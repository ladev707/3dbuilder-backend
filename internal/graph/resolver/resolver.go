// Package resolver answers GraphQL queries and mutations.
//
// Resolvers are a thin adapter: they turn GraphQL arguments into service
// calls and service results into GraphQL return values. Business rules live
// in the services, so the same rules apply to any other API added later.
//
// File layout (set by `resolver.layout: follow-schema` in gqlgen.yml): each
// schema file gets a matching resolver file, e.g. schema/user.graphqls maps
// to user.resolvers.go. Running `make graphql` adds stubs for new fields and
// keeps the bodies you have written. This file is never regenerated, so it
// holds dependencies and shared helpers.
package resolver

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/ladev707/3dbuilder-backend/internal/apperr"
	authmiddleware "github.com/ladev707/3dbuilder-backend/internal/modules/auth/middleware"
	customerservice "github.com/ladev707/3dbuilder-backend/internal/modules/customer/service"
	userservice "github.com/ladev707/3dbuilder-backend/internal/modules/user/service"
)

// Resolver is the root resolver and holds the dependencies every resolver
// method can use (r.UserService, r.CustomerService). It is created once in
// internal/api/app.go. To expose a new module, add its service here.
type Resolver struct {
	UserService     *userservice.Service
	CustomerService *customerservice.Service
}

// isCurrentPrincipal reports whether id belongs to the caller. The JWT
// "subject" claim holds the caller's ID.
func isCurrentPrincipal(ctx context.Context, id uuid.UUID) bool {
	claims, ok := authmiddleware.ClaimsFromRequestContext(ctx)
	return ok && claims.Subject == id.String()
}

// isNotFound reports whether a service returned an apperr NotFound error.
// Single-record queries use it to return null instead of an error.
func isNotFound(err error) bool {
	var appErr *apperr.Error
	return errors.As(err, &appErr) && appErr.Kind == apperr.NotFound
}

// valueOrZero dereferences an optional argument, returning T's zero value
// when it is nil. gqlgen passes optional GraphQL arguments as pointers (nil
// means "not provided").
//
// It is generic ([T any]), so the same function works for *int, *UserFilter
// and any other pointer type.
func valueOrZero[T any](value *T) T {
	if value == nil {
		var zero T
		return zero
	}
	return *value
}
