-- +goose Up
CREATE TYPE property_status AS ENUM ('active', 'inactive');

CREATE TABLE properties (
    id                  BIGSERIAL       PRIMARY KEY,
    name                VARCHAR(200)    NOT NULL,
    address             TEXT            NOT NULL,
    area                NUMERIC(10, 2),
    unit_count          INTEGER,
    status              property_status NOT NULL DEFAULT 'active',
    management_company  VARCHAR(200),
    assignee_id         BIGINT          REFERENCES users (id) ON DELETE SET NULL,
    created_by          BIGINT          REFERENCES users (id) ON DELETE SET NULL,
    updated_by          BIGINT          REFERENCES users (id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_properties_status      ON properties (status);
CREATE INDEX idx_properties_assignee_id ON properties (assignee_id);
CREATE INDEX idx_properties_name        ON properties (name);

-- +goose Down
DROP TABLE IF EXISTS properties;
DROP TYPE IF EXISTS property_status;
