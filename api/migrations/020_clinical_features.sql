-- 020_clinical_features.sql
-- New clinical domains + platform features

-- ============================================================
-- Immunization Domain
-- ============================================================

CREATE TABLE IF NOT EXISTS immunization (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    patient_id UUID NOT NULL,
    encounter_id UUID,
    vaccine_code_system VARCHAR(255),
    vaccine_code VARCHAR(20) NOT NULL,
    vaccine_display VARCHAR(500) NOT NULL,
    occurrence_datetime TIMESTAMPTZ,
    occurrence_string VARCHAR(255),
    primary_source BOOLEAN NOT NULL DEFAULT true,
    lot_number VARCHAR(100),
    expiration_date DATE,
    site_code VARCHAR(20),
    site_display VARCHAR(255),
    route_code VARCHAR(20),
    route_display VARCHAR(255),
    dose_quantity NUMERIC(10,2),
    dose_unit VARCHAR(20),
    performer_id UUID,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS immunization_recommendation (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) NOT NULL UNIQUE,
    patient_id UUID NOT NULL,
    date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    vaccine_code VARCHAR(20) NOT NULL,
    vaccine_display VARCHAR(500) NOT NULL,
    forecast_status VARCHAR(50) NOT NULL,
    forecast_display VARCHAR(255),
    date_criterion TIMESTAMPTZ,
    series_doses INTEGER,
    dose_number INTEGER,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- CarePlan Domain
-- ============================================================

CREATE TABLE IF NOT EXISTS care_plan (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) NOT NULL UNIQUE,
    status VARCHAR(30) NOT NULL DEFAULT 'draft',
    intent VARCHAR(20) NOT NULL DEFAULT 'plan',
    category_code VARCHAR(50),
    category_display VARCHAR(255),
    title VARCHAR(500),
    description TEXT,
    patient_id UUID NOT NULL,
    encounter_id UUID,
    period_start TIMESTAMPTZ,
    period_end TIMESTAMPTZ,
    author_id UUID,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS care_plan_activity (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    care_plan_id UUID NOT NULL REFERENCES care_plan(id) ON DELETE CASCADE,
    detail_code VARCHAR(50),
    detail_display VARCHAR(255),
    status VARCHAR(30) NOT NULL DEFAULT 'not-started',
    scheduled_start TIMESTAMPTZ,
    scheduled_end TIMESTAMPTZ,
    description TEXT
);

CREATE TABLE IF NOT EXISTS goal (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) NOT NULL UNIQUE,
    lifecycle_status VARCHAR(30) NOT NULL DEFAULT 'proposed',
    achievement_status VARCHAR(50),
    category_code VARCHAR(50),
    category_display VARCHAR(255),
    description TEXT NOT NULL,
    patient_id UUID NOT NULL,
    target_measure VARCHAR(255),
    target_detail_string VARCHAR(500),
    target_due_date DATE,
    expressed_by_id UUID,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- FamilyHistory Domain
-- ============================================================

CREATE TABLE IF NOT EXISTS family_member_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) NOT NULL UNIQUE,
    status VARCHAR(30) NOT NULL DEFAULT 'completed',
    patient_id UUID NOT NULL,
    date TIMESTAMPTZ,
    name VARCHAR(255),
    relationship_code VARCHAR(50) NOT NULL,
    relationship_display VARCHAR(255) NOT NULL,
    sex VARCHAR(20),
    born_date DATE,
    deceased_boolean BOOLEAN,
    deceased_age INTEGER,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS family_member_condition (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    family_member_id UUID NOT NULL REFERENCES family_member_history(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    display VARCHAR(255) NOT NULL,
    outcome_code VARCHAR(50),
    outcome_display VARCHAR(255),
    contributed_to_death BOOLEAN,
    onset_age INTEGER
);

-- ============================================================
-- RelatedPerson Domain
-- ============================================================

CREATE TABLE IF NOT EXISTS related_person (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) NOT NULL UNIQUE,
    active BOOLEAN NOT NULL DEFAULT true,
    patient_id UUID NOT NULL,
    relationship_code VARCHAR(50) NOT NULL,
    relationship_display VARCHAR(255) NOT NULL,
    family_name VARCHAR(255),
    given_name VARCHAR(255),
    phone VARCHAR(50),
    email VARCHAR(255),
    gender VARCHAR(20),
    birth_date DATE,
    address_line VARCHAR(500),
    address_city VARCHAR(100),
    address_state VARCHAR(50),
    address_postal_code VARCHAR(20),
    period_start TIMESTAMPTZ,
    period_end TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS related_person_communication (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    related_person_id UUID NOT NULL REFERENCES related_person(id) ON DELETE CASCADE,
    language_code VARCHAR(20) NOT NULL,
    language_display VARCHAR(100) NOT NULL,
    preferred BOOLEAN NOT NULL DEFAULT false
);

-- ============================================================
-- Provenance Domain
-- ============================================================

CREATE TABLE IF NOT EXISTS provenance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) NOT NULL UNIQUE,
    target_type VARCHAR(100) NOT NULL,
    target_id VARCHAR(64) NOT NULL,
    recorded TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    activity_code VARCHAR(50),
    activity_display VARCHAR(255),
    reason_code VARCHAR(50),
    reason_display VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS provenance_agent (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provenance_id UUID NOT NULL REFERENCES provenance(id) ON DELETE CASCADE,
    type_code VARCHAR(50),
    type_display VARCHAR(255),
    who_type VARCHAR(100) NOT NULL,
    who_id VARCHAR(64) NOT NULL,
    on_behalf_of_type VARCHAR(100),
    on_behalf_of_id VARCHAR(64)
);

CREATE TABLE IF NOT EXISTS provenance_entity (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provenance_id UUID NOT NULL REFERENCES provenance(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL,
    what_type VARCHAR(100) NOT NULL,
    what_id VARCHAR(64) NOT NULL
);

-- ============================================================
-- Patient MPI (PatientLink) - extend existing table from 001
-- ============================================================

-- patient_link already exists from 001_t0_core_tables.sql with columns:
--   id, patient_id, linked_patient_id, link_type, created_at
-- Add new columns for the MPI feature.
ALTER TABLE patient_link ADD COLUMN IF NOT EXISTS confidence NUMERIC(5,4) DEFAULT 0;
ALTER TABLE patient_link ADD COLUMN IF NOT EXISTS match_method VARCHAR(50) DEFAULT '';
ALTER TABLE patient_link ADD COLUMN IF NOT EXISTS match_details TEXT;
ALTER TABLE patient_link ADD COLUMN IF NOT EXISTS created_by VARCHAR(255) DEFAULT '';

-- ============================================================
-- Order Status History
-- ============================================================

CREATE TABLE IF NOT EXISTS order_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID NOT NULL,
    from_status VARCHAR(30) NOT NULL,
    to_status VARCHAR(30) NOT NULL,
    changed_by VARCHAR(255),
    reason TEXT,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- Document Templates
-- ============================================================

CREATE TABLE IF NOT EXISTS document_template (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    sections JSONB,
    variables JSONB,
    author_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS template_section (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES document_template(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    order_index INTEGER NOT NULL DEFAULT 0,
    content_template TEXT NOT NULL,
    required BOOLEAN NOT NULL DEFAULT false,
    variables JSONB
);

-- ============================================================
-- Export Jobs
-- ============================================================

CREATE TABLE IF NOT EXISTS export_job (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status VARCHAR(20) NOT NULL DEFAULT 'accepted',
    resource_types TEXT[],
    since TIMESTAMPTZ,
    output_format VARCHAR(50) NOT NULL DEFAULT 'application/fhir+ndjson',
    output_files JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    error_message TEXT
);

-- ============================================================
-- Measure Reports (Reporting)
-- ============================================================

CREATE TABLE IF NOT EXISTS measure_report (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    measure_id VARCHAR(100) NOT NULL,
    measure_name VARCHAR(255) NOT NULL,
    period_start DATE,
    period_end DATE,
    results JSONB,
    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- Performance Indexes
-- ============================================================

-- Indexes for new tables
CREATE INDEX IF NOT EXISTS idx_immunization_patient ON immunization(patient_id);
CREATE INDEX IF NOT EXISTS idx_immunization_patient_status ON immunization(patient_id, status, created_at);
CREATE INDEX IF NOT EXISTS idx_immunization_recommendation_patient ON immunization_recommendation(patient_id);
CREATE INDEX IF NOT EXISTS idx_care_plan_patient ON care_plan(patient_id);
CREATE INDEX IF NOT EXISTS idx_care_plan_patient_status ON care_plan(patient_id, status, created_at);
CREATE INDEX IF NOT EXISTS idx_goal_patient ON goal(patient_id);
CREATE INDEX IF NOT EXISTS idx_family_member_history_patient ON family_member_history(patient_id);
CREATE INDEX IF NOT EXISTS idx_related_person_patient ON related_person(patient_id);
CREATE INDEX IF NOT EXISTS idx_provenance_target ON provenance(target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_patient_link_source ON patient_link(patient_id);
CREATE INDEX IF NOT EXISTS idx_patient_link_target ON patient_link(linked_patient_id);
CREATE INDEX IF NOT EXISTS idx_order_status_history_resource ON order_status_history(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_document_template_type ON document_template(type, status);

-- Performance indexes for existing tables
CREATE INDEX IF NOT EXISTS idx_condition_patient_status ON condition(patient_id, clinical_status, created_at);
CREATE INDEX IF NOT EXISTS idx_observation_patient_status ON observation(patient_id, status, created_at);
CREATE INDEX IF NOT EXISTS idx_observation_patient_code ON observation(patient_id, code_value);
CREATE INDEX IF NOT EXISTS idx_allergy_patient ON allergy_intolerance(patient_id, created_at);
CREATE INDEX IF NOT EXISTS idx_procedure_patient_status ON procedure_record(patient_id, status, created_at);
CREATE INDEX IF NOT EXISTS idx_encounter_patient_status ON encounter(patient_id, status, created_at);
CREATE INDEX IF NOT EXISTS idx_medication_request_patient ON medication_request(patient_id, status, created_at);
CREATE INDEX IF NOT EXISTS idx_service_request_patient ON service_request(patient_id, status, created_at);
CREATE INDEX IF NOT EXISTS idx_diagnostic_report_patient ON diagnostic_report(patient_id, status, created_at);
CREATE INDEX IF NOT EXISTS idx_appointment_patient ON appointment(patient_id, status);
CREATE INDEX IF NOT EXISTS idx_consent_patient ON consent(patient_id, status);
CREATE INDEX IF NOT EXISTS idx_document_ref_patient ON document_reference(patient_id, status);
