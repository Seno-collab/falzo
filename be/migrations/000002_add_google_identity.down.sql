DROP TABLE IF EXISTS auth_identities;

ALTER TABLE users
ALTER COLUMN password_hash SET NOT NULL;
