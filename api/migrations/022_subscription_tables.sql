-- 022_subscription_tables.sql
-- FHIR R4 Subscription support with REST-hook webhook delivery.
-- Enables real-time notifications when resources change.

CREATE TABLE IF NOT EXISTS subscription (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'requested',
    criteria        TEXT NOT NULL,
    channel_type    TEXT NOT NULL DEFAULT 'rest-hook',
    channel_endpoint TEXT NOT NULL,
    channel_payload TEXT NOT NULL DEFAULT 'application/fhir+json',
    channel_headers JSONB,
    end_time        TIMESTAMPTZ,
    error_text      TEXT,
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_subscription_status ON subscription (status);
CREATE INDEX IF NOT EXISTS idx_subscription_criteria ON subscription (criteria);

CREATE TABLE IF NOT EXISTS subscription_notification (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID NOT NULL REFERENCES subscription(id) ON DELETE CASCADE,
    resource_type   TEXT NOT NULL,
    resource_id     TEXT NOT NULL,
    event_type      TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending',
    payload         JSONB,
    attempt_count   INTEGER NOT NULL DEFAULT 0,
    max_attempts    INTEGER NOT NULL DEFAULT 5,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_error      TEXT,
    delivered_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_subscription_notification_status ON subscription_notification (status);
CREATE INDEX IF NOT EXISTS idx_subscription_notification_next_attempt ON subscription_notification (next_attempt_at)
    WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_subscription_notification_subscription_id ON subscription_notification (subscription_id);
