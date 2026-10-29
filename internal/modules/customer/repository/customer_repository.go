// Package customerrepository is the only place that runs SQL against the
// customers table. It follows the same patterns as userrepository; see the
// comments there for the reasoning behind COALESCE updates, NULL-able filters
// and soft deletes.
package customerrepository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ladev707/3dbuilder-backend/internal/database"
	customermodel "github.com/ladev707/3dbuilder-backend/internal/modules/customer/models"
)

// Sentinel errors returned by the repository and mapped by the service.
var (
	ErrCustomerNotFound  = errors.New("customer not found")
	ErrDuplicateCustomer = errors.New("username or email already in use")
	// ErrRoleNotFound means the role passed to Create is missing from the
	// roles table, i.e. the seed data in the migrations was not applied.
	ErrRoleNotFound = errors.New("default customer role not found")
)

// customerColumns must stay in the same order as the Scan in scanCustomer.
const customerColumns = `id, username, email, password_hash, is_active, created_at, updated_at`

// Repository stores customers in Postgres.
type Repository struct {
	db *pgxpool.Pool
}

// New creates a Repository that uses the shared connection pool.
func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// List returns one page of non-deleted customers matching params.Filter,
// plus the total number of matches.
func (r *Repository) List(ctx context.Context, params customermodel.ListCustomersParams) ([]customermodel.Customer, int, error) {
	search := database.ContainsPattern(params.Filter.Search)
	// A NULL parameter disables its filter ("$1::text IS NULL" is true).
	const where = `
		WHERE deleted_at IS NULL
		  AND ($1::text IS NULL OR username ILIKE $1 OR email ILIKE $1)
		  AND ($2::boolean IS NULL OR is_active = $2)`

	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM customers`+where, search, params.Filter.IsActive).Scan(&total); err != nil {
		return nil, 0, err
	}

	// "id" breaks ties so the order is stable across pages.
	rows, err := r.db.Query(ctx, `SELECT `+customerColumns+` FROM customers`+where+`
		ORDER BY created_at DESC, id
		LIMIT $3 OFFSET $4`, search, params.Filter.IsActive, params.Limit, params.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	customers := []customermodel.Customer{}
	for rows.Next() {
		customer, err := scanCustomer(rows)
		if err != nil {
			return nil, 0, err
		}
		customers = append(customers, customer)
	}
	return customers, total, rows.Err()
}

// FindByID returns a non-deleted customer, or ErrCustomerNotFound.
func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (customermodel.Customer, error) {
	row := r.db.QueryRow(ctx, `SELECT `+customerColumns+` FROM customers WHERE id = $1 AND deleted_at IS NULL`, id)
	return scanCustomer(row)
}

// Create inserts the customer and assigns defaultRole in one transaction.
//
// Without a role the customer could log in but would have no permissions
// (for example /auth/customer/me requires "profile.read"). The transaction
// guarantees we never end up with a customer that has no role.
func (r *Repository) Create(ctx context.Context, customer customermodel.Customer, defaultRole string) (customermodel.Customer, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return customermodel.Customer{}, err
	}
	defer tx.Rollback(ctx)

	customer, err = scanCustomer(tx.QueryRow(ctx, `
		INSERT INTO customers (username, email, password_hash, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING `+customerColumns, customer.Username, customer.Email, customer.PasswordHash, customer.IsActive))
	if err != nil {
		return customermodel.Customer{}, err
	}
	// INSERT ... SELECT looks up the role ID by name. If the role does not
	// exist the SELECT returns no rows, so nothing is inserted.
	tag, err := tx.Exec(ctx, `
		INSERT INTO customer_roles (customer_id, role_id, gate)
		SELECT $1, id, gate FROM roles WHERE gate = 'customer' AND name = $2`, customer.ID, defaultRole)
	if err != nil {
		return customermodel.Customer{}, err
	}
	if tag.RowsAffected() != 1 {
		return customermodel.Customer{}, ErrRoleNotFound
	}
	return customer, tx.Commit(ctx)
}

// Update applies the non-nil fields of changes (COALESCE keeps the current
// value for NULL parameters). It revokes the customer's sessions when the
// password changes or the account ends up inactive.
func (r *Repository) Update(ctx context.Context, id uuid.UUID, changes customermodel.CustomerChanges) (customermodel.Customer, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return customermodel.Customer{}, err
	}
	defer tx.Rollback(ctx)

	customer, err := scanCustomer(tx.QueryRow(ctx, `
		UPDATE customers SET
			username = COALESCE($2, username),
			email = COALESCE($3, email),
			password_hash = COALESCE($4, password_hash),
			is_active = COALESCE($5, is_active),
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING `+customerColumns, id, changes.Username, changes.Email, changes.PasswordHash, changes.IsActive))
	if err != nil {
		return customermodel.Customer{}, err
	}
	if changes.PasswordHash != nil || !customer.IsActive {
		if err := revokeSessions(ctx, tx, id); err != nil {
			return customermodel.Customer{}, err
		}
	}
	return customer, tx.Commit(ctx)
}

// Delete soft-deletes the customer (sets deleted_at) so the username and
// email can be reused, and revokes their sessions.
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
		UPDATE customers SET deleted_at = NOW(), is_active = FALSE, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrCustomerNotFound
	}
	if err := revokeSessions(ctx, tx, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// revokeSessions invalidates every active refresh token of the customer.
func revokeSessions(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		UPDATE auth_sessions SET revoked_at = NOW()
		WHERE gate = 'customer' AND principal_id = $1 AND revoked_at IS NULL`, id)
	return err
}

// scanCustomer reads one row in customerColumns order and translates
// "no rows" and unique violations into repository errors.
func scanCustomer(row pgx.Row) (customermodel.Customer, error) {
	var customer customermodel.Customer
	err := row.Scan(
		&customer.ID,
		&customer.Username,
		&customer.Email,
		&customer.PasswordHash,
		&customer.IsActive,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return customermodel.Customer{}, ErrCustomerNotFound
	case database.IsUniqueViolation(err):
		return customermodel.Customer{}, ErrDuplicateCustomer
	}
	return customer, err
}
