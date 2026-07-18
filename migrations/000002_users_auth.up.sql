CREATE TYPE user_role   AS ENUM ('user', 'admin');
CREATE TYPE user_status AS ENUM ('active', 'suspended', 'deleted');

CREATE TABLE users (
    id                bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email             text        NOT NULL UNIQUE,
    username          text        UNIQUE,
    password_hash     text,                 -- NULL for SSO-only accounts (Phase 2)
    full_name         text        NOT NULL,
    phone             text,
    whatsapp_number   text,
    role              user_role   NOT NULL DEFAULT 'user',
    status            user_status NOT NULL DEFAULT 'active',
    email_verified_at timestamptz,
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_users_updated BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Phase-2 SSO: external identities linked to a user. Local password auth stays on users.
CREATE TABLE auth_providers (
    id               bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id          bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider         text   NOT NULL,          -- 'google', 'apple', ...
    provider_user_id text   NOT NULL,
    created_at       timestamptz NOT NULL DEFAULT now(),
    UNIQUE (provider, provider_user_id)
);
CREATE INDEX idx_auth_providers_user ON auth_providers(user_id);

-- Refresh tokens (access tokens are stateless JWT). Stored hashed for revocation.
CREATE TABLE sessions (
    id                 bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id            bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash text   NOT NULL UNIQUE,
    user_agent         text,
    ip_address         inet,
    expires_at         timestamptz NOT NULL,
    revoked_at         timestamptz,
    created_at         timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_sessions_user ON sessions(user_id);
