// Package graph builds the HTTP handler that serves the GraphQL API.
//
// How the pieces fit together:
//
//	schema/*.graphqls     the API contract (edit these)
//	generated/            code produced by `make graphql` (never edit)
//	resolver/             your code that answers each query and mutation
//	handler.go (this)     wires the generated schema to the resolvers and adds
//	                      the permission directive, error formatting and
//	                      safety limits
//
// The handler is mounted in internal/api/v1/routes.go behind the Authenticate
// and RequireGate(user) middleware, so every request reaching it already has
// a valid back-office access token.
package graph

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/ladev707/3dbuilder-backend/internal/apperr"
	"github.com/ladev707/3dbuilder-backend/internal/graph/generated"
	"github.com/ladev707/3dbuilder-backend/internal/graph/resolver"
	authmiddleware "github.com/ladev707/3dbuilder-backend/internal/modules/auth/middleware"
	"github.com/ladev707/3dbuilder-backend/pkg/logger"
)

// maxQueryComplexity caps how much work one request may ask for. By default
// every field costs 1, so a deeply nested or very wide query is rejected
// before it runs. This protects the server from denial-of-service queries.
const maxQueryComplexity = 200

// Options configures NewHandler.
type Options struct {
	// Introspection lets tools like the Playground discover the schema.
	// Disable it in production so outsiders cannot map the whole API.
	Introspection bool
}

// NewHandler returns an http.Handler that executes GraphQL requests using
// resolvers.
func NewHandler(resolvers *resolver.Resolver, options Options) http.Handler {
	// NewExecutableSchema (generated code) connects every schema field to a
	// resolver method. Directives maps each `@directive` declared in the
	// schema to the Go function that implements it.
	schema := generated.NewExecutableSchema(generated.Config{
		Resolvers: resolvers,
		Directives: generated.DirectiveRoot{
			HasPermission: hasPermission,
		},
	})

	// handler.New starts with no transports, so we opt in to each one:
	// OPTIONS for CORS preflight, GET for queries in the URL, and POST with a
	// JSON body (what most clients use). Mutations are never allowed over GET.
	server := handler.New(schema)
	server.AddTransport(transport.Options{})
	server.AddTransport(transport.GET{})
	server.AddTransport(transport.POST{})
	// Parsed and validated query documents are cached, so repeated queries
	// skip parsing.
	server.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	server.Use(extension.FixedComplexityLimit(maxQueryComplexity))
	if options.Introspection {
		server.Use(extension.Introspection{})
	}
	server.SetErrorPresenter(presentError)
	// A panic inside a resolver is caught here instead of crashing the
	// server. The returned error then goes through presentError.
	server.SetRecoverFunc(func(ctx context.Context, recovered any) error {
		logger.Error("graphql resolver panicked", "panic", fmt.Sprint(recovered))
		return errors.New("internal server error")
	})
	return server
}

// hasPermission implements the schema directive
//
//	directive @hasPermission(permission: String!) on FIELD_DEFINITION
//
// gqlgen calls it before the resolver of every field annotated with
// @hasPermission. next runs the real resolver, so returning an error without
// calling next blocks the field. It is the per-field equivalent of the
// RequirePermission Gin middleware used by the REST routes.
func hasPermission(ctx context.Context, _ any, next graphql.Resolver, permission string) (any, error) {
	// The Authenticate middleware stored the JWT claims in the request
	// context; see authmiddleware.ContextWithClaims.
	claims, ok := authmiddleware.ClaimsFromRequestContext(ctx)
	if !ok {
		return nil, apperr.New(apperr.Unauthenticated, "authentication required")
	}
	// Permissions are copied into the access token at login, so this check
	// needs no database query.
	if !slices.Contains(claims.Permissions, permission) {
		return nil, apperr.New(apperr.Forbidden, "required permission is missing: "+permission)
	}
	return next(ctx)
}

// presentError decides what a client sees for each resolver error.
//
//   - *apperr.Error: its message and kind are shown, e.g.
//     {"message":"user not found","extensions":{"code":"NOT_FOUND"}}
//   - *gqlerror.Error (raised by gqlgen itself, e.g. an invalid UUID
//     argument): shown unchanged.
//   - anything else (database failures, bugs): logged with its real cause,
//     and the client only gets "internal server error", so internal details
//     such as SQL never leak.
func presentError(ctx context.Context, err error) *gqlerror.Error {
	// DefaultErrorPresenter adds the "path" (which field failed) and
	// "locations" that clients expect.
	presented := graphql.DefaultErrorPresenter(ctx, err)

	var appErr *apperr.Error
	if errors.As(err, &appErr) {
		presented.Message = appErr.Message
		presented.Extensions = map[string]any{"code": string(appErr.Kind)}
		return presented
	}
	var gqlErr *gqlerror.Error
	if errors.As(err, &gqlErr) {
		return presented
	}

	logger.Error("graphql resolver failed", "error", err, "path", presented.Path.String())
	presented.Message = "internal server error"
	presented.Extensions = map[string]any{"code": "INTERNAL_SERVER_ERROR"}
	return presented
}
