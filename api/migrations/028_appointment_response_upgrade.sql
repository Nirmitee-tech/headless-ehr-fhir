-- Upgrade appointment_response to a full FHIR resource
ALTER TABLE appointment_response ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE appointment_response ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();
