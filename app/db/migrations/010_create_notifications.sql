-- +goose Up
CREATE TABLE notifications (
    id          BIGSERIAL    PRIMARY KEY,
    user_id     BIGINT       NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    title       VARCHAR(200) NOT NULL,
    body        TEXT         NOT NULL,
    link        VARCHAR(500),
    is_read     BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_id  ON notifications (user_id);
CREATE INDEX idx_notifications_is_read  ON notifications (user_id, is_read);
CREATE INDEX idx_notifications_created  ON notifications (created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS notifications;
