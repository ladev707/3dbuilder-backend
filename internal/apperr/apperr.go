package apperr

// Kind classifies an error. Its value is sent to clients as
// extensions.code, so frontends can branch on it. Never rename existing values.
type Kind string

const (
	// InvalidInput means the request data failed validation.
	InvalidInput Kind = "BAD_USER_INPUT"
	// NotFound means the requested record does not exist (or was soft-deleted).
	NotFound Kind = "NOT_FOUND"
	// Conflict means the request clashes with existing data, e.g. a duplicate email.
	Conflict Kind = "CONFLICT"
	// Unauthenticated means no valid access token was provided.
	Unauthenticated Kind = "UNAUTHENTICATED"
	// Forbidden means the caller is authenticated but not allowed to do this.
	Forbidden Kind = "FORBIDDEN"
)

// Error is an error whose Message is safe to return to API clients.
type Error struct {
	Kind    Kind
	Message string
}

// Error implements the built-in error interface, so *Error can be returned
// anywhere a normal error is expected.
func (e *Error) Error() string {
	return e.Message
}

// New creates a client-safe error.
//
//	return apperr.New(apperr.NotFound, "user not found")
func New(kind Kind, message string) *Error {
	return &Error{Kind: kind, Message: message}
}
