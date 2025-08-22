CREATE SCHEMA IF NOT EXISTS provider;

CREATE TABLE IF NOT EXISTS provider.processed_events (
  payment_id    TEXT PRIMARY KEY,
  processed_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  status        TEXT NOT NULL,                  -- AUTHORIZED | DECLINED
  psp_reference TEXT NULL
);