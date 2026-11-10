-- +goose Up
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

-- +goose Down
DROP TABLE IF EXISTS customers;
