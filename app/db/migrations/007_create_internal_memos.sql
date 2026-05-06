-- +goose Up
CREATE TABLE internal_memos (
    id          BIGSERIAL   PRIMARY KEY,
    claim_id    BIGINT      NOT NULL REFERENCES claims (id) ON DELETE CASCADE,
    content     TEXT        NOT NULL,
    created_by  BIGINT      REFERENCES users (id) ON DELETE SET NULL,
    updated_by  BIGINT      REFERENCES users (id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_internal_memos_claim_id   ON internal_memos (claim_id);
CREATE INDEX idx_internal_memos_created_by ON internal_memos (created_by);

-- +goose Down
DROP TABLE IF EXISTS internal_memos;
