ALTER TABLE users
ALTER COLUMN password_hash DROP NOT NULL;

CREATE TABLE auth_identities (
	id BIGSERIAL PRIMARY KEY,
	user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	provider VARCHAR(30) NOT NULL,
	provider_subject VARCHAR(255) NOT NULL,
	provider_email VARCHAR(255),
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	UNIQUE (provider, provider_subject)
);
