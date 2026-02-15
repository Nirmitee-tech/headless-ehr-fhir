-- Upgrade practitioner_role to a full FHIR resource
ALTER TABLE practitioner_role ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE practitioner_role ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();
