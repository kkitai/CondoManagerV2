-- +goose Up
CREATE TABLE attachments (
    id           BIGSERIAL    PRIMARY KEY,
    claim_id     BIGINT       NOT NULL REFERENCES claims (id) ON DELETE CASCADE,
    file_name    VARCHAR(255) NOT NULL,
    file_path    TEXT         NOT NULL,
    file_size    BIGINT       NOT NULL,
    mime_type    VARCHAR(100) NOT NULL,
    uploaded_by  BIGINT       REFERENCES users (id) ON DELETE SET NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_attachments_claim_id    ON attachments (claim_id);
CREATE INDEX idx_attachments_uploaded_by ON attachments (uploaded_by);

-- +goose Down
DROP TABLE IF EXISTS attachments;
