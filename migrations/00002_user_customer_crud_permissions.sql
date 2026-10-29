-- Permissions for the GraphQL user and customer CRUD API.
--
-- 00001 already seeds users.read and users.create. This adds the rest and
-- grants them all to the back-office "admin" role. Each name matches a
-- @hasPermission(permission: "...") in internal/graph/schema/*.graphqls.
--
-- Permissions are copied into the access token at login, so existing admins
-- must log in again (or refresh) to receive them.

-- +goose Up
-- ON CONFLICT DO NOTHING makes the migration safe to run against a database
-- where some of these rows were created by hand.
INSERT INTO permissions (gate, name, description) VALUES
    ('user', 'users.update', 'Update users'),
    ('user', 'users.delete', 'Delete users'),
    ('user', 'customers.read', 'View customers'),
    ('user', 'customers.create', 'Create customers'),
    ('user', 'customers.update', 'Update customers'),
    ('user', 'customers.delete', 'Delete customers')
ON CONFLICT (gate, name) DO NOTHING;

-- Grant the new permissions to the admin role. INSERT ... SELECT looks up
-- the IDs by name, so no UUIDs are hard-coded.
INSERT INTO role_permissions (role_id, permission_id, gate)
SELECT r.id, p.id, r.gate
FROM roles r
JOIN permissions p ON p.gate = r.gate
WHERE r.gate = 'user'
  AND r.name = 'admin'
  AND p.name IN ('users.update', 'users.delete', 'customers.read', 'customers.create', 'customers.update', 'customers.delete')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- +goose Down
-- role_permissions rows are removed automatically through ON DELETE CASCADE.
DELETE FROM permissions
WHERE gate = 'user'
  AND name IN ('users.update', 'users.delete', 'customers.read', 'customers.create', 'customers.update', 'customers.delete');
