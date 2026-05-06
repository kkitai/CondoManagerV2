-- +goose Up
CREATE TYPE user_role AS ENUM ('admin', 'general');
CREATE TYPE user_status AS ENUM ('active', 'invited', 'disabled');

CREATE TABLE users (
    id              BIGSERIAL PRIMARY KEY,
    email           VARCHAR(255) NOT NULL UNIQUE,
    password_hash   VARCHAR(255),
    name            VARCHAR(100) NOT NULL,
    role            user_role    NOT NULL DEFAULT 'general',
    department      VARCHAR(100),
    job_title       VARCHAR(100),
    status          user_status  NOT NULL DEFAULT 'invited',
    invited_at      TIMESTAMPTZ,
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email  ON users (email);
CREATE INDEX idx_users_status ON users (status);
CREATE INDEX idx_users_role   ON users (role);

-- +goose Down
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS user_status;
DROP TYPE IF EXISTS user_role;
