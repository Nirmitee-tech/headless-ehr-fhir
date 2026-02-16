-- 026_group_table.sql
-- Creates the fhir_group and fhir_group_member tables for the FHIR Group resource.
-- Table name avoids SQL reserved keyword "group".

CREATE TABLE IF NOT EXISTS fhir_group (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id           TEXT NOT NULL UNIQUE,
    group_type        TEXT NOT NULL DEFAULT 'person',
    actual            BOOLEAN NOT NULL DEFAULT true,
    active            BOOLEAN NOT NULL DEFAULT true,
    name              TEXT NOT NULL,
    code              TEXT,
    quantity          INTEGER NOT NULL DEFAULT 0,
    managing_entity   TEXT,
    version_id        INTEGER NOT NULL DEFAULT 1,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_fhir_group_fhir_id ON fhir_group (fhir_id);
CREATE INDEX IF NOT EXISTS idx_fhir_group_type ON fhir_group (group_type);
CREATE INDEX IF NOT EXISTS idx_fhir_group_name ON fhir_group (name);
CREATE INDEX IF NOT EXISTS idx_fhir_group_active ON fhir_group (active);

CREATE TABLE IF NOT EXISTS fhir_group_member (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id          UUID NOT NULL REFERENCES fhir_group(id) ON DELETE CASCADE,
    entity_type       TEXT NOT NULL,
    entity_id         TEXT NOT NULL,
    period_start      TIMESTAMPTZ,
    period_end        TIMESTAMPTZ,
    inactive          BOOLEAN NOT NULL DEFAULT false,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_fhir_group_member_group ON fhir_group_member (group_id);
CREATE INDEX IF NOT EXISTS idx_fhir_group_member_entity ON fhir_group_member (entity_type, entity_id);
