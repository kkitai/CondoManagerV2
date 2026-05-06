-- +goose Up
CREATE TABLE tags (
    id          BIGSERIAL    PRIMARY KEY,
    name        VARCHAR(100) NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE claim_tags (
    claim_id  BIGINT NOT NULL REFERENCES claims (id) ON DELETE CASCADE,
    tag_id    BIGINT NOT NULL REFERENCES tags (id) ON DELETE CASCADE,
    PRIMARY KEY (claim_id, tag_id)
);

CREATE INDEX idx_claim_tags_tag_id ON claim_tags (tag_id);

-- +goose Down
DROP TABLE IF EXISTS claim_tags;
DROP TABLE IF EXISTS tags;
