-- Migration 019: Add resource versioning support
-- Adds version_id column to all resource tables and creates shared resource_history table

-- Shared history table for FHIR resource versioning (vread / _history)
CREATE TABLE IF NOT EXISTS resource_history (
    id              BIGSERIAL PRIMARY KEY,
    resource_type   VARCHAR(64) NOT NULL,
    resource_id     VARCHAR(64) NOT NULL,
    version_id      INTEGER NOT NULL,
    resource        JSONB NOT NULL,
    action          VARCHAR(16) NOT NULL DEFAULT 'update', -- create, update, delete
    timestamp       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_resource_history_lookup
    ON resource_history (resource_type, resource_id, version_id);

CREATE INDEX IF NOT EXISTS idx_resource_history_type_id
    ON resource_history (resource_type, resource_id, timestamp DESC);

-- Add version_id to all resource tables
-- Identity domain
ALTER TABLE patient ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE practitioner ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;

-- Admin domain
ALTER TABLE organization ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE department ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE location ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;

-- Encounter domain
ALTER TABLE encounter ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;

-- Clinical domain
ALTER TABLE condition ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE observation ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE allergy_intolerance ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE procedure_record ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;

-- Medication domain
ALTER TABLE medication ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE medication_request ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE medication_administration ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE medication_dispense ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE medication_statement ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;

-- Diagnostics domain
ALTER TABLE service_request ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE specimen ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE diagnostic_report ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE imaging_study ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;

-- Scheduling domain
ALTER TABLE schedule ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE slot ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE appointment ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;

-- Billing domain
ALTER TABLE coverage ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE claim ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE claim_response ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE invoice ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;

-- Documents domain
ALTER TABLE consent ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE document_reference ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE clinical_note ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE composition ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;

-- Inbox domain
ALTER TABLE message_pool ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE inbox_message ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;

-- Surgery domain
ALTER TABLE or_room ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE surgical_case ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;

-- Emergency domain
ALTER TABLE triage_record ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE ed_tracking ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE trauma_activation ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;

-- Obstetrics domain
ALTER TABLE pregnancy ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE prenatal_visit ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE labor_record ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE delivery_record ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE newborn_record ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE postpartum_record ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;

-- Oncology domain
ALTER TABLE cancer_diagnosis ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE treatment_protocol ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE chemotherapy_cycle ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE radiation_therapy ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE tumor_marker ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE tumor_board_review ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;

-- Nursing domain
ALTER TABLE flowsheet_template ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE flowsheet_entry ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE nursing_assessment ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;

-- Behavioral domain
ALTER TABLE psychiatric_assessment ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE safety_plan ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE legal_hold ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;

-- Research domain
ALTER TABLE research_study ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE research_enrollment ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;

-- Portal domain
ALTER TABLE questionnaire ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE questionnaire_response ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;

-- CDS domain
ALTER TABLE cds_rule ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE cds_alert ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;

