-- 038: Add care_team tables and provenance version_id
-- Required for Inferno g(10) US Core CareTeam and Provenance support.

-- ============================================================
-- CareTeam
-- ============================================================
CREATE TABLE IF NOT EXISTS care_team (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    name VARCHAR(255),
    patient_id UUID NOT NULL REFERENCES patient(id),
    encounter_id UUID REFERENCES encounter(id),
    category_code VARCHAR(100),
    category_display VARCHAR(255),
    period_start DATE,
    period_end DATE,
    managing_organization_id UUID REFERENCES organization(id),
    reason_code VARCHAR(100),
    reason_display VARCHAR(255),
    note TEXT,
    version_id INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_care_team_fhir_id ON care_team(fhir_id);
CREATE INDEX IF NOT EXISTS idx_care_team_patient_id ON care_team(patient_id);

CREATE TABLE IF NOT EXISTS care_team_participant (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    care_team_id UUID NOT NULL REFERENCES care_team(id) ON DELETE CASCADE,
    member_id UUID NOT NULL,
    member_type VARCHAR(50) NOT NULL,
    role_code VARCHAR(50),
    role_display VARCHAR(255),
    period_start DATE,
    period_end DATE,
    on_behalf_of_id UUID
);

CREATE INDEX IF NOT EXISTS idx_care_team_participant_care_team_id ON care_team_participant(care_team_id);

-- ============================================================
-- Provenance: add version_id column (missing from original migration)
-- ============================================================
ALTER TABLE provenance ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
