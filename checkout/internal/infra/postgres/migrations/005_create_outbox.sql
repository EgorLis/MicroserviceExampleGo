CREATE TABLE IF NOT EXISTS checkout.outbox_events (
    id              BIGSERIAL PRIMARY KEY,
    aggregate_type  TEXT        NOT NULL,                 -- "payment"
    aggregate_id    TEXT        NOT NULL,                 -- payment_id
    event_type      TEXT        NOT NULL,                 -- "payment.created"
    event_version   INT         NOT NULL DEFAULT 1,
    key             TEXT        NOT NULL,                 -- партиционирование
    payload         JSONB       NOT NULL,                 -- value (готовый JSON)
    headers         JSONB       NOT NULL DEFAULT '{}'::jsonb,
    status          TEXT        NOT NULL DEFAULT 'NEW',   -- NEW|SENT|FAILED
    attempt         INT         NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS outbox_status_idx ON checkout.outbox_events (status, next_attempt_at);
CREATE INDEX IF NOT EXISTS outbox_created_idx ON checkout.outbox_events (created_at);