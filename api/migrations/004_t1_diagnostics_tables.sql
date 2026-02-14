-- ============================================================================
-- T1 DIAGNOSTICS TABLES MIGRATION
-- Tables: service_request, specimen, diagnostic_report,
--         diagnostic_report_result, imaging_study
-- ============================================================================

-- ============================================================================
-- 1. SERVICE_REQUEST (from 10_diagnostics_lab.sql)
-- ============================================================================

CREATE TABLE service_request (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id                 VARCHAR(64) UNIQUE NOT NULL,
    patient_id              UUID NOT NULL REFERENCES patient(id),
    encounter_id            UUID REFERENCES encounter(id),
    requester_id            UUID NOT NULL REFERENCES practitioner(id),
    performer_id            UUID REFERENCES practitioner(id),
    status                  VARCHAR(20) NOT NULL,
    intent                  VARCHAR(20) NOT NULL,
    priority                VARCHAR(20),
    category_code           VARCHAR(30),
    category_display        VARCHAR(255),
    code_system             VARCHAR(255),
    code_value              VARCHAR(30) NOT NULL,
    code_display            VARCHAR(500) NOT NULL,
    order_detail_code       VARCHAR(30),
    order_detail_display    VARCHAR(255),
    quantity_value          DECIMAL(12,4),
    quantity_unit           VARCHAR(50),
    occurrence_datetime     TIMESTAMPTZ,
    occurrence_start        TIMESTAMPTZ,
    occurrence_end          TIMESTAMPTZ,
    authored_on             TIMESTAMPTZ DEFAULT NOW(),
    reason_code             VARCHAR(30),
    reason_display          VARCHAR(255),
    reason_condition_id     UUID REFERENCES condition(id),
    specimen_requirement    TEXT,
    body_site_code          VARCHAR(30),
    body_site_display       VARCHAR(255),
    note                    TEXT,
    patient_instruction     TEXT,
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 2. SPECIMEN (from 10_diagnostics_lab.sql)
-- ============================================================================

CREATE TABLE specimen (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id                 VARCHAR(64) UNIQUE NOT NULL,
    patient_id              UUID NOT NULL REFERENCES patient(id),
    accession_id            VARCHAR(100),
    status                  VARCHAR(20) NOT NULL,
    type_code               VARCHAR(30),
    type_display            VARCHAR(255),
    received_time           TIMESTAMPTZ,
    collection_collector    UUID REFERENCES practitioner(id),
    collection_datetime     TIMESTAMPTZ,
    collection_quantity     DECIMAL(12,4),
    collection_unit         VARCHAR(50),
    collection_method       VARCHAR(30),
    collection_body_site    VARCHAR(30),
    processing_description  TEXT,
    processing_procedure    VARCHAR(30),
    processing_datetime     TIMESTAMPTZ,
    container_description   VARCHAR(255),
    container_type          VARCHAR(30),
    condition_code          VARCHAR(30),
    condition_display       VARCHAR(255),
    note                    TEXT,
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 3. DIAGNOSTIC_REPORT (from 10_diagnostics_lab.sql)
-- ============================================================================

CREATE TABLE diagnostic_report (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id                 VARCHAR(64) UNIQUE NOT NULL,
    patient_id              UUID NOT NULL REFERENCES patient(id),
    encounter_id            UUID REFERENCES encounter(id),
    performer_id            UUID REFERENCES practitioner(id),
    status                  VARCHAR(20) NOT NULL,
    category_code           VARCHAR(30),
    category_display        VARCHAR(255),
    code_system             VARCHAR(255),
    code_value              VARCHAR(30) NOT NULL,
    code_display            VARCHAR(500) NOT NULL,
    effective_datetime      TIMESTAMPTZ,
    effective_start         TIMESTAMPTZ,
    effective_end           TIMESTAMPTZ,
    issued                  TIMESTAMPTZ DEFAULT NOW(),
    specimen_id             UUID REFERENCES specimen(id),
    conclusion              TEXT,
    conclusion_code         VARCHAR(30),
    conclusion_display      VARCHAR(255),
    presented_form_url      TEXT,
    presented_form_type     VARCHAR(100),
    note                    TEXT,
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 4. DIAGNOSTIC_REPORT_RESULT (junction table)
-- ============================================================================

CREATE TABLE diagnostic_report_result (
    diagnostic_report_id    UUID NOT NULL REFERENCES diagnostic_report(id) ON DELETE CASCADE,
    observation_id          UUID NOT NULL REFERENCES observation(id) ON DELETE CASCADE,
    PRIMARY KEY (diagnostic_report_id, observation_id)
);

-- ============================================================================
-- 5. IMAGING_STUDY (from 10_diagnostics_lab.sql)
-- ============================================================================

CREATE TABLE imaging_study (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id                 VARCHAR(64) UNIQUE NOT NULL,
    patient_id              UUID NOT NULL REFERENCES patient(id),
    encounter_id            UUID REFERENCES encounter(id),
    referrer_id             UUID REFERENCES practitioner(id),
    status                  VARCHAR(20) NOT NULL,
    modality_code           VARCHAR(10),
    modality_display        VARCHAR(100),
    study_uid               VARCHAR(255),
    number_of_series        INTEGER,
    number_of_instances     INTEGER,
    description             TEXT,
    started                 TIMESTAMPTZ,
    endpoint                VARCHAR(500),
    reason_code             VARCHAR(30),
    reason_display          VARCHAR(255),
    note                    TEXT,
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 6. INDEXES for T1 diagnostics tables
-- ============================================================================

-- ServiceRequest indexes
CREATE INDEX idx_service_request_fhir ON service_request(fhir_id);
CREATE INDEX idx_service_request_patient ON service_request(patient_id);
CREATE INDEX idx_service_request_encounter ON service_request(encounter_id) WHERE encounter_id IS NOT NULL;
CREATE INDEX idx_service_request_requester ON service_request(requester_id);
CREATE INDEX idx_service_request_status ON service_request(status);
CREATE INDEX idx_service_request_code ON service_request(code_value);
CREATE INDEX idx_service_request_category ON service_request(category_code) WHERE category_code IS NOT NULL;
CREATE INDEX idx_service_request_authored ON service_request(authored_on DESC) WHERE authored_on IS NOT NULL;

-- Specimen indexes
CREATE INDEX idx_specimen_fhir ON specimen(fhir_id);
CREATE INDEX idx_specimen_patient ON specimen(patient_id);
CREATE INDEX idx_specimen_status ON specimen(status);
CREATE INDEX idx_specimen_type ON specimen(type_code) WHERE type_code IS NOT NULL;
CREATE INDEX idx_specimen_accession ON specimen(accession_id) WHERE accession_id IS NOT NULL;

-- DiagnosticReport indexes
CREATE INDEX idx_diagnostic_report_fhir ON diagnostic_report(fhir_id);
CREATE INDEX idx_diagnostic_report_patient ON diagnostic_report(patient_id);
CREATE INDEX idx_diagnostic_report_encounter ON diagnostic_report(encounter_id) WHERE encounter_id IS NOT NULL;
CREATE INDEX idx_diagnostic_report_status ON diagnostic_report(status);
CREATE INDEX idx_diagnostic_report_code ON diagnostic_report(code_value);
CREATE INDEX idx_diagnostic_report_category ON diagnostic_report(category_code) WHERE category_code IS NOT NULL;
CREATE INDEX idx_diagnostic_report_issued ON diagnostic_report(issued DESC) WHERE issued IS NOT NULL;
CREATE INDEX idx_diagnostic_report_specimen ON diagnostic_report(specimen_id) WHERE specimen_id IS NOT NULL;

-- DiagnosticReportResult indexes
CREATE INDEX idx_dr_result_observation ON diagnostic_report_result(observation_id);

-- ImagingStudy indexes
CREATE INDEX idx_imaging_study_fhir ON imaging_study(fhir_id);
CREATE INDEX idx_imaging_study_patient ON imaging_study(patient_id);
CREATE INDEX idx_imaging_study_encounter ON imaging_study(encounter_id) WHERE encounter_id IS NOT NULL;
CREATE INDEX idx_imaging_study_status ON imaging_study(status);
CREATE INDEX idx_imaging_study_modality ON imaging_study(modality_code) WHERE modality_code IS NOT NULL;
CREATE INDEX idx_imaging_study_started ON imaging_study(started DESC) WHERE started IS NOT NULL;
CREATE INDEX idx_imaging_study_uid ON imaging_study(study_uid) WHERE study_uid IS NOT NULL;
