-- +goose Up
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

-- +goose Down
DROP TABLE IF EXISTS auth_sessions;
