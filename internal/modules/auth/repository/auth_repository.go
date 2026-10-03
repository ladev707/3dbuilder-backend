package authrepository

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	authmodel "github.com/ladev707/3dbuilder-backend/internal/src/auth/models"
)

var (
	ErrPrincipalNotFound = errors.New("principal not found")
	ErrInvalidSession    = errors.New("invalid or expired session")
	ErrInvalidAssignment = errors.New("role or permission does not belong to the selected gate")
)

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindPrincipalByEmail(ctx context.Context, gate authmodel.Gate, email string) (authmodel.Principal, error) {
	table, err := principalTable(gate)
	if err != nil {
		return authmodel.Principal{}, err
	}
	query := fmt.Sprintf(`
		SELECT id, username, email, password_hash, is_active
		FROM %s
		WHERE LOWER(email) = LOWER($1) AND deleted_at IS NULL`, table)

	principal := authmodel.Principal{Gate: gate}
	err = r.db.QueryRow(ctx, query, email).Scan(
		&principal.ID,
		&principal.Username,
		&principal.Email,
		&principal.PasswordHash,
		&principal.Active,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return authmodel.Principal{}, ErrPrincipalNotFound
	}
	return principal, err
}

func (r *Repository) FindPrincipalByID(ctx context.Context, gate authmodel.Gate, id uuid.UUID) (authmodel.Principal, error) {
	table, err := principalTable(gate)
	if err != nil {
		return authmodel.Principal{}, err
	}
	query := fmt.Sprintf(`
		SELECT id, username, email, password_hash, is_active
		FROM %s
		WHERE id = $1 AND deleted_at IS NULL`, table)

	principal := authmodel.Principal{Gate: gate}
	err = r.db.QueryRow(ctx, query, id).Scan(
		&principal.ID,
		&principal.Username,
		&principal.Email,
		&principal.PasswordHash,
		&principal.Active,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return authmodel.Principal{}, ErrPrincipalNotFound
	}
	return principal, err
}

func (r *Repository) LoadAuthorization(ctx context.Context, principal *authmodel.Principal) error {
	joinTable, ownerColumn, err := authorizationTable(principal.Gate)
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`
		SELECT r.name, COALESCE(ARRAY_AGG(DISTINCT p.name) FILTER (WHERE p.name IS NOT NULL), '{}')
		FROM %s ar
		JOIN roles r ON r.id = ar.role_id AND r.gate = ar.gate
		LEFT JOIN role_permissions rp ON rp.role_id = r.id AND rp.gate = r.gate
		LEFT JOIN permissions p ON p.id = rp.permission_id AND p.gate = r.gate
		WHERE ar.%s = $1
		GROUP BY r.name
		ORDER BY r.name`, joinTable, ownerColumn)

	rows, err := r.db.Query(ctx, query, principal.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	principal.Roles = []string{}
	principal.Permissions = []string{}
	permissionSet := make(map[string]struct{})
	for rows.Next() {
		var role string
		var permissions []string
		if err := rows.Scan(&role, &permissions); err != nil {
			return err
		}
		principal.Roles = append(principal.Roles, role)
		for _, permission := range permissions {
			permissionSet[permission] = struct{}{}
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for permission := range permissionSet {
		principal.Permissions = append(principal.Permissions, permission)
	}
	sort.Strings(principal.Permissions)
	return nil
}

func (r *Repository) ListRoles(ctx context.Context, gate authmodel.Gate) ([]authmodel.Role, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, gate, name, description
		FROM roles WHERE gate = $1 ORDER BY name`, gate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	roles := []authmodel.Role{}
	for rows.Next() {
		var role authmodel.Role
		if err := rows.Scan(&role.ID, &role.Gate, &role.Name, &role.Description); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *Repository) ListPermissions(ctx context.Context, gate authmodel.Gate) ([]authmodel.Permission, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, gate, name, description
		FROM permissions WHERE gate = $1 ORDER BY name`, gate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	permissions := []authmodel.Permission{}
	for rows.Next() {
		var permission authmodel.Permission
		if err := rows.Scan(&permission.ID, &permission.Gate, &permission.Name, &permission.Description); err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	return permissions, rows.Err()
}

func (r *Repository) CreateRole(ctx context.Context, gate authmodel.Gate, name, description string) (authmodel.Role, error) {
	role := authmodel.Role{Gate: gate, Name: name, Description: description}
	err := r.db.QueryRow(ctx, `
		INSERT INTO roles (gate, name, description)
		VALUES ($1, $2, $3)
		RETURNING id`, gate, name, description).Scan(&role.ID)
	return role, err
}

func (r *Repository) AssignPermission(ctx context.Context, gate authmodel.Gate, roleID, permissionID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
		INSERT INTO role_permissions (role_id, permission_id, gate)
		SELECT r.id, p.id, $1::varchar
		FROM roles r, permissions p
		WHERE r.id = $2 AND r.gate = $1::varchar AND p.id = $3 AND p.gate = $1::varchar
		ON CONFLICT (role_id, permission_id) DO UPDATE SET gate = EXCLUDED.gate`, gate, roleID, permissionID)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrInvalidAssignment
	}
	return err
}

func (r *Repository) RemovePermission(ctx context.Context, gate authmodel.Gate, roleID, permissionID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM role_permissions
		WHERE gate = $1 AND role_id = $2 AND permission_id = $3`, gate, roleID, permissionID)
	return err
}

func (r *Repository) AssignRole(ctx context.Context, gate authmodel.Gate, principalID, roleID uuid.UUID) error {
	table, ownerColumn, err := authorizationTable(gate)
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`
		INSERT INTO %s (%s, role_id, gate)
		SELECT $1, id, $2::varchar FROM roles WHERE id = $3 AND gate = $2::varchar
		ON CONFLICT (%s, role_id) DO UPDATE SET gate = EXCLUDED.gate`, table, ownerColumn, ownerColumn)
	tag, err := r.db.Exec(ctx, query, principalID, gate, roleID)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrInvalidAssignment
	}
	return err
}

func (r *Repository) RemoveRole(ctx context.Context, gate authmodel.Gate, principalID, roleID uuid.UUID) error {
	table, ownerColumn, err := authorizationTable(gate)
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`DELETE FROM %s WHERE %s = $1 AND role_id = $2 AND gate = $3`, table, ownerColumn)
	_, err = r.db.Exec(ctx, query, principalID, roleID, gate)
	return err
}

func (r *Repository) CreateSession(ctx context.Context, session authmodel.Session) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO auth_sessions (id, principal_id, gate, refresh_token_hash, expires_at)
		VALUES ($1, $2, $3, $4, $5)`,
		session.ID, session.PrincipalID, session.Gate, session.RefreshTokenHash, session.ExpiresAt,
	)
	return err
}

func (r *Repository) RotateSession(ctx context.Context, oldID uuid.UUID, oldTokenHash string, next authmodel.Session) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
		UPDATE auth_sessions
		SET revoked_at = NOW()
		WHERE id = $1
		  AND refresh_token_hash = $2
		  AND revoked_at IS NULL
		  AND expires_at > NOW()`, oldID, oldTokenHash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrInvalidSession
	}
	if _, err = tx.Exec(ctx, `
		INSERT INTO auth_sessions (id, principal_id, gate, refresh_token_hash, expires_at)
		VALUES ($1, $2, $3, $4, $5)`,
		next.ID, next.PrincipalID, next.Gate, next.RefreshTokenHash, next.ExpiresAt,
	); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) RevokeSession(ctx context.Context, sessionID, principalID uuid.UUID, gate authmodel.Gate) error {
	_, err := r.db.Exec(ctx, `
		UPDATE auth_sessions SET revoked_at = COALESCE(revoked_at, NOW())
		WHERE id = $1 AND principal_id = $2 AND gate = $3`,
		sessionID, principalID, gate,
	)
	return err
}

func principalTable(gate authmodel.Gate) (string, error) {
	switch gate {
	case authmodel.GateUser:
		return "users", nil
	case authmodel.GateCustomer:
		return "customers", nil
	default:
		return "", fmt.Errorf("invalid authentication gate %q", gate)
	}
}

func authorizationTable(gate authmodel.Gate) (string, string, error) {
	switch gate {
	case authmodel.GateUser:
		return "user_roles", "user_id", nil
	case authmodel.GateCustomer:
		return "customer_roles", "customer_id", nil
	default:
		return "", "", fmt.Errorf("invalid authentication gate %q", gate)
	}
}
