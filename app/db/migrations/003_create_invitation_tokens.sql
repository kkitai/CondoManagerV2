-- +goose Up
CREATE TABLE invitation_tokens (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT       NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash  VARCHAR(255) NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ  NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invitation_tokens_token_hash ON invitation_tokens (token_hash);
CREATE INDEX idx_invitation_tokens_user_id    ON invitation_tokens (user_id);

-- +goose Down
DROP TABLE IF EXISTS invitation_tokens;
