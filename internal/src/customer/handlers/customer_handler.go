package customerhandler

// Handler is the HTTP adapter for customer-management use cases.
// Authentication endpoints live in the auth module.
type Handler struct{}

func New() *Handler {
	return &Handler{}
}
