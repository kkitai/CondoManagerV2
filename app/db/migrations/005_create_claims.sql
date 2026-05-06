-- +goose Up
CREATE TYPE claim_status AS ENUM ('pending', 'in_progress', 'completed');
CREATE TYPE claim_severity AS ENUM ('low', 'medium', 'high', 'urgent');

CREATE TABLE claims (
    id                  BIGSERIAL      PRIMARY KEY,
    title               VARCHAR(300)   NOT NULL,
    content             TEXT           NOT NULL,
    property_id         BIGINT         NOT NULL REFERENCES properties (id) ON DELETE RESTRICT,
    reporter_name       VARCHAR(100)   NOT NULL,
    reporter_contact    VARCHAR(255),
    status              claim_status   NOT NULL DEFAULT 'pending',
    severity            claim_severity NOT NULL DEFAULT 'medium',
    category            VARCHAR(100),
    assignee_id         BIGINT         REFERENCES users (id) ON DELETE SET NULL,
    is_recurrence       BOOLEAN        NOT NULL DEFAULT FALSE,
    response_due_at     TIMESTAMPTZ,
    completed_at        TIMESTAMPTZ,
    satisfaction_score  SMALLINT       CHECK (satisfaction_score BETWEEN 1 AND 5),
    created_by          BIGINT         REFERENCES users (id) ON DELETE SET NULL,
    updated_by          BIGINT         REFERENCES users (id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_claims_property_id   ON claims (property_id);
CREATE INDEX idx_claims_status        ON claims (status);
CREATE INDEX idx_claims_severity      ON claims (severity);
CREATE INDEX idx_claims_assignee_id   ON claims (assignee_id);
CREATE INDEX idx_claims_category      ON claims (category);
CREATE INDEX idx_claims_created_at    ON claims (created_at DESC);
CREATE INDEX idx_claims_fulltext      ON claims USING GIN (to_tsvector('japanese', title || ' ' || content));

-- +goose Down
DROP TABLE IF EXISTS claims;
DROP TYPE IF EXISTS claim_severity;
DROP TYPE IF EXISTS claim_status;
