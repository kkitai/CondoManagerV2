-- +goose Up
CREATE TYPE response_type AS ENUM ('comment', 'response_history', 'system_log');

CREATE TABLE claim_responses (
    id          BIGSERIAL     PRIMARY KEY,
    claim_id    BIGINT        NOT NULL REFERENCES claims (id) ON DELETE CASCADE,
    type        response_type NOT NULL,
    content     TEXT          NOT NULL,
    created_by  BIGINT        REFERENCES users (id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_claim_responses_claim_id   ON claim_responses (claim_id);
CREATE INDEX idx_claim_responses_type       ON claim_responses (type);
CREATE INDEX idx_claim_responses_created_by ON claim_responses (created_by);

-- +goose Down
DROP TABLE IF EXISTS claim_responses;
DROP TYPE IF EXISTS response_type;
