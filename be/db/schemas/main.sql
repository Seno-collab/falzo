CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    username      VARCHAR(50)  NOT NULL UNIQUE,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name     VARCHAR(255) NULL,
    is_active     BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS roles (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(50)  NOT NULL UNIQUE,
    description VARCHAR(255) NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_roles (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL,
    role_id     BIGINT NOT NULL,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_user_role UNIQUE (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id           BIGSERIAL PRIMARY KEY,
    session_id   VARCHAR(64)  UNIQUE,
    user_id      BIGINT NOT NULL,
    username     VARCHAR(50)  NULL,
    subject      VARCHAR(255) NULL,
    token_hash   VARCHAR(255) NOT NULL UNIQUE,
    device_info  VARCHAR(255) NULL,
    ip_address   VARCHAR(45)  NULL,
    is_revoked   BOOLEAN      NOT NULL DEFAULT FALSE,
    expires_at   TIMESTAMPTZ  NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

ALTER TABLE refresh_tokens
    ADD COLUMN IF NOT EXISTS session_id VARCHAR(64);

ALTER TABLE refresh_tokens
    ADD COLUMN IF NOT EXISTS username VARCHAR(50);

ALTER TABLE refresh_tokens
    ADD COLUMN IF NOT EXISTS subject VARCHAR(255);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id
    ON refresh_tokens (user_id);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at
    ON refresh_tokens (expires_at);

CREATE UNIQUE INDEX IF NOT EXISTS uq_refresh_tokens_session_id
    ON refresh_tokens (session_id);

-- Seed data mac dinh
INSERT INTO roles (name, description) VALUES
  ('admin',     'Full system access'),
  ('moderator', 'Content moderation'),
  ('user',      'Standard user')
ON CONFLICT (name) DO UPDATE
SET description = EXCLUDED.description;
