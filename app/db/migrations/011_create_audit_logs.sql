-- +goose Up
CREATE TABLE audit_logs (
    id           BIGSERIAL    PRIMARY KEY,
    user_id      BIGINT       REFERENCES users (id) ON DELETE SET NULL,
    action       VARCHAR(100) NOT NULL,
    entity_type  VARCHAR(100) NOT NULL,
    entity_id    BIGINT       NOT NULL,
    before_state JSONB,
    after_state  JSONB,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user_id     ON audit_logs (user_id);
CREATE INDEX idx_audit_logs_entity      ON audit_logs (entity_type, entity_id);
CREATE INDEX idx_audit_logs_created_at  ON audit_logs (created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS audit_logs;
