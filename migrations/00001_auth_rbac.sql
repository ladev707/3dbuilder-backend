-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(80) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX users_username_unique ON users (LOWER(username)) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX users_email_unique ON users (LOWER(email)) WHERE deleted_at IS NULL;

CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(80) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX customers_username_unique ON customers (LOWER(username)) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX customers_email_unique ON customers (LOWER(email)) WHERE deleted_at IS NULL;

CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    gate VARCHAR(20) NOT NULL CHECK (gate IN ('user', 'customer')),
    name VARCHAR(80) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (id, gate),
    UNIQUE (gate, name)
);

CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    gate VARCHAR(20) NOT NULL CHECK (gate IN ('user', 'customer')),
    name VARCHAR(120) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (id, gate),
    UNIQUE (gate, name)
);

CREATE TABLE role_permissions (
    role_id UUID NOT NULL,
    permission_id UUID NOT NULL,
    gate VARCHAR(20) NOT NULL CHECK (gate IN ('user', 'customer')),
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id, gate) REFERENCES roles (id, gate) ON DELETE CASCADE,
    FOREIGN KEY (permission_id, gate) REFERENCES permissions (id, gate) ON DELETE CASCADE
);

CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role_id UUID NOT NULL,
    gate VARCHAR(20) NOT NULL DEFAULT 'user' CHECK (gate = 'user'),
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (role_id, gate) REFERENCES roles (id, gate) ON DELETE CASCADE
);

CREATE TABLE customer_roles (
    customer_id UUID NOT NULL REFERENCES customers (id) ON DELETE CASCADE,
    role_id UUID NOT NULL,
    gate VARCHAR(20) NOT NULL DEFAULT 'customer' CHECK (gate = 'customer'),
    PRIMARY KEY (customer_id, role_id),
    FOREIGN KEY (role_id, gate) REFERENCES roles (id, gate) ON DELETE CASCADE
);

CREATE TABLE auth_sessions (
    id UUID PRIMARY KEY,
    principal_id UUID NOT NULL,
    gate VARCHAR(20) NOT NULL CHECK (gate IN ('user', 'customer')),
    refresh_token_hash CHAR(64) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX auth_sessions_principal_idx ON auth_sessions (gate, principal_id);
CREATE INDEX auth_sessions_expiry_idx ON auth_sessions (expires_at);

INSERT INTO roles (gate, name, description) VALUES
    ('user', 'admin', 'Full back-office access'),
    ('user', 'member', 'Standard back-office user'),
    ('customer', 'customer', 'Standard customer');

INSERT INTO permissions (gate, name, description) VALUES
    ('user', 'users.read', 'View users'),
    ('user', 'users.create', 'Create users'),
    ('user', 'roles.manage', 'Manage roles and permissions'),
    ('customer', 'profile.read', 'View own customer profile'),
    ('customer', 'profile.update', 'Update own customer profile');

INSERT INTO role_permissions (role_id, permission_id, gate)
SELECT r.id, p.id, r.gate
FROM roles r
JOIN permissions p ON p.gate = r.gate
WHERE r.name IN ('admin', 'customer');

-- +goose Down
DROP TABLE IF EXISTS auth_sessions;
DROP TABLE IF EXISTS customer_roles;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS customers;
DROP TABLE IF EXISTS users;
