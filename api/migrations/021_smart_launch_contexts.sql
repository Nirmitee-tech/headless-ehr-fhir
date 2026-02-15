-- 021_smart_launch_contexts.sql
-- Persistent storage for SMART on FHIR launch contexts.
-- Replaces the in-memory store to support horizontal scaling.

CREATE TABLE IF NOT EXISTS smart_launch_contexts (
    id          TEXT PRIMARY KEY,
    context_json JSONB NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at  TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_smart_launch_contexts_expires_at
    ON smart_launch_contexts (expires_at);
