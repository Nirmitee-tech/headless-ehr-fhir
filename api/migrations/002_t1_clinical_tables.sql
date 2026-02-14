-- ============================================================================
-- T1 CLINICAL TABLES MIGRATION
-- Tables: condition, observation, observation_component, allergy_intolerance,
--         allergy_reaction, procedure_record, procedure_performer,
--         procedure_focal_device
-- ============================================================================

-- ============================================================================
-- 1. CONDITION (from 05_conditions_diagnoses.sql)
-- ============================================================================

CREATE TABLE condition (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    patient_id          UUID NOT NULL REFERENCES patient(id),
    encounter_id        UUID REFERENCES encounter(id),
    recorder_id         UUID REFERENCES practitioner(id),
    asserter_id         UUID REFERENCES practitioner(id),
    clinical_status     VARCHAR(20) NOT NULL,
    verification_status VARCHAR(20),
    category_code       VARCHAR(30),
    severity_code       VARCHAR(20),
    severity_display    VARCHAR(50),
    code_system         VARCHAR(255),
    code_value          VARCHAR(30) NOT NULL,
    code_display        VARCHAR(500) NOT NULL,
    alt_code_system     VARCHAR(255),
    alt_code_value      VARCHAR(30),
    alt_code_display    VARCHAR(500),
    body_site_code      VARCHAR(30),
    body_site_display   VARCHAR(255),
    onset_datetime      TIMESTAMPTZ,
    onset_age           INTEGER,
    onset_period_start  TIMESTAMPTZ,
    onset_period_end    TIMESTAMPTZ,
    onset_string        VARCHAR(255),
    abatement_datetime  TIMESTAMPTZ,
    abatement_age       INTEGER,
    abatement_period_start TIMESTAMPTZ,
    abatement_period_end   TIMESTAMPTZ,
    abatement_string    VARCHAR(255),
    stage_summary_code  VARCHAR(30),
    stage_summary_display VARCHAR(255),
    stage_type_code     VARCHAR(30),
    evidence_code       VARCHAR(30),
    evidence_display    VARCHAR(255),
    recorded_date       TIMESTAMPTZ DEFAULT NOW(),
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE encounter_diagnosis ADD CONSTRAINT fk_enc_diag_condition
    FOREIGN KEY (condition_id) REFERENCES condition(id);

-- ============================================================================
-- 2. OBSERVATION (from 06_observations_vitals.sql)
-- ============================================================================

CREATE TABLE observation (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    status              VARCHAR(20) NOT NULL,
    category_code       VARCHAR(50),
    category_display    VARCHAR(100),
    code_system         VARCHAR(255),
    code_value          VARCHAR(30) NOT NULL,
    code_display        VARCHAR(500) NOT NULL,
    patient_id          UUID NOT NULL REFERENCES patient(id),
    encounter_id        UUID REFERENCES encounter(id),
    performer_id        UUID REFERENCES practitioner(id),
    effective_datetime  TIMESTAMPTZ,
    effective_period_start TIMESTAMPTZ,
    effective_period_end   TIMESTAMPTZ,
    issued              TIMESTAMPTZ DEFAULT NOW(),
    value_quantity      DECIMAL(12,4),
    value_unit          VARCHAR(50),
    value_system        VARCHAR(255),
    value_code          VARCHAR(20),
    value_string        TEXT,
    value_boolean       BOOLEAN,
    value_integer       INTEGER,
    value_date_time     TIMESTAMPTZ,
    value_codeable_code VARCHAR(30),
    value_codeable_display VARCHAR(255),
    value_ratio_numerator   DECIMAL(12,4),
    value_ratio_denominator DECIMAL(12,4),
    reference_range_low     DECIMAL(12,4),
    reference_range_high    DECIMAL(12,4),
    reference_range_unit    VARCHAR(50),
    reference_range_text    VARCHAR(255),
    reference_range_age_low  INTEGER,
    reference_range_age_high INTEGER,
    interpretation_code     VARCHAR(20),
    interpretation_display  VARCHAR(100),
    body_site_code      VARCHAR(30),
    body_site_display   VARCHAR(255),
    method_code         VARCHAR(30),
    method_display      VARCHAR(255),
    device_id           UUID,
    specimen_id         UUID,
    data_absent_reason  VARCHAR(30),
    note                TEXT,
    has_member          BOOLEAN DEFAULT FALSE,
    derived_from_id     UUID REFERENCES observation(id),
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE observation_component (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    observation_id      UUID NOT NULL REFERENCES observation(id) ON DELETE CASCADE,
    code_system         VARCHAR(255),
    code_value          VARCHAR(30) NOT NULL,
    code_display        VARCHAR(255) NOT NULL,
    value_quantity      DECIMAL(12,4),
    value_unit          VARCHAR(50),
    value_string        TEXT,
    value_codeable_code VARCHAR(30),
    value_codeable_display VARCHAR(255),
    interpretation_code VARCHAR(20),
    interpretation_display VARCHAR(100),
    reference_range_low  DECIMAL(12,4),
    reference_range_high DECIMAL(12,4),
    reference_range_unit VARCHAR(50),
    reference_range_text VARCHAR(255)
);

-- ============================================================================
-- 3. ALLERGY INTOLERANCE (from 07_allergies.sql)
-- ============================================================================

CREATE TABLE allergy_intolerance (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    patient_id          UUID NOT NULL REFERENCES patient(id),
    encounter_id        UUID REFERENCES encounter(id),
    recorder_id         UUID REFERENCES practitioner(id),
    asserter_id         UUID REFERENCES practitioner(id),
    clinical_status     VARCHAR(20),
    verification_status VARCHAR(20),
    type                VARCHAR(20),
    category            VARCHAR(20)[],
    criticality         VARCHAR(20),
    code_system         VARCHAR(255),
    code_value          VARCHAR(30),
    code_display        VARCHAR(500),
    onset_datetime      TIMESTAMPTZ,
    onset_age           INTEGER,
    onset_string        VARCHAR(255),
    recorded_date       TIMESTAMPTZ DEFAULT NOW(),
    last_occurrence     TIMESTAMPTZ,
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE allergy_reaction (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    allergy_id          UUID NOT NULL REFERENCES allergy_intolerance(id) ON DELETE CASCADE,
    substance_code      VARCHAR(30),
    substance_display   VARCHAR(255),
    manifestation_code  VARCHAR(30) NOT NULL,
    manifestation_display VARCHAR(255) NOT NULL,
    description         TEXT,
    severity            VARCHAR(20),
    exposure_route_code VARCHAR(20),
    exposure_route_display VARCHAR(100),
    onset               TIMESTAMPTZ,
    note                TEXT
);

-- ============================================================================
-- 4. PROCEDURE (from 09_procedures.sql)
-- ============================================================================

CREATE TABLE procedure_record (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    status              VARCHAR(30) NOT NULL,
    status_reason_code  VARCHAR(30),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    encounter_id        UUID REFERENCES encounter(id),
    recorder_id         UUID REFERENCES practitioner(id),
    asserter_id         UUID REFERENCES practitioner(id),
    code_system         VARCHAR(255),
    code_value          VARCHAR(30) NOT NULL,
    code_display        VARCHAR(500) NOT NULL,
    category_code       VARCHAR(30),
    category_display    VARCHAR(255),
    performed_datetime  TIMESTAMPTZ,
    performed_start     TIMESTAMPTZ,
    performed_end       TIMESTAMPTZ,
    performed_string    VARCHAR(255),
    body_site_code      VARCHAR(30),
    body_site_display   VARCHAR(255),
    outcome_code        VARCHAR(30),
    outcome_display     VARCHAR(255),
    complication_code   VARCHAR(30),
    complication_display VARCHAR(255),
    follow_up_code      VARCHAR(30),
    follow_up_display   VARCHAR(255),
    reason_code         VARCHAR(30),
    reason_display      VARCHAR(255),
    reason_condition_id UUID REFERENCES condition(id),
    report_id           UUID,
    location_id         UUID REFERENCES location(id),
    used_reference_text TEXT,
    nabh_procedure_code VARCHAR(30),
    cpt_code            VARCHAR(10),
    cpt_modifier        VARCHAR(10)[],
    hcpcs_code          VARCHAR(10),
    anesthesia_type     VARCHAR(30),
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE procedure_performer (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    procedure_id        UUID NOT NULL REFERENCES procedure_record(id) ON DELETE CASCADE,
    practitioner_id     UUID NOT NULL REFERENCES practitioner(id),
    role_code           VARCHAR(30),
    role_display        VARCHAR(100),
    organization_id     UUID REFERENCES organization(id),
    on_behalf_of_org_id UUID REFERENCES organization(id)
);

CREATE TABLE procedure_focal_device (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    procedure_id        UUID NOT NULL REFERENCES procedure_record(id) ON DELETE CASCADE,
    action_code         VARCHAR(20),
    device_id           UUID
);

-- ============================================================================
-- 5. INDEXES for T1 tables
-- ============================================================================

-- Condition indexes
CREATE INDEX idx_condition_fhir ON condition(fhir_id);
CREATE INDEX idx_condition_patient ON condition(patient_id);
CREATE INDEX idx_condition_encounter ON condition(encounter_id) WHERE encounter_id IS NOT NULL;
CREATE INDEX idx_condition_code ON condition(code_value);
CREATE INDEX idx_condition_status ON condition(clinical_status);
CREATE INDEX idx_condition_recorded ON condition(recorded_date DESC);

-- Observation indexes
CREATE INDEX idx_observation_fhir ON observation(fhir_id);
CREATE INDEX idx_observation_patient ON observation(patient_id);
CREATE INDEX idx_observation_encounter ON observation(encounter_id) WHERE encounter_id IS NOT NULL;
CREATE INDEX idx_observation_code ON observation(code_value);
CREATE INDEX idx_observation_category ON observation(category_code) WHERE category_code IS NOT NULL;
CREATE INDEX idx_observation_status ON observation(status);
CREATE INDEX idx_observation_effective ON observation(effective_datetime DESC) WHERE effective_datetime IS NOT NULL;
CREATE INDEX idx_observation_patient_code ON observation(patient_id, code_value);

-- Allergy indexes
CREATE INDEX idx_allergy_fhir ON allergy_intolerance(fhir_id);
CREATE INDEX idx_allergy_patient ON allergy_intolerance(patient_id);
CREATE INDEX idx_allergy_status ON allergy_intolerance(clinical_status) WHERE clinical_status IS NOT NULL;
CREATE INDEX idx_allergy_criticality ON allergy_intolerance(criticality) WHERE criticality IS NOT NULL;

-- Procedure indexes
CREATE INDEX idx_procedure_fhir ON procedure_record(fhir_id);
CREATE INDEX idx_procedure_patient ON procedure_record(patient_id);
CREATE INDEX idx_procedure_encounter ON procedure_record(encounter_id) WHERE encounter_id IS NOT NULL;
CREATE INDEX idx_procedure_code ON procedure_record(code_value);
CREATE INDEX idx_procedure_status ON procedure_record(status);
CREATE INDEX idx_procedure_date ON procedure_record(performed_datetime DESC) WHERE performed_datetime IS NOT NULL;

-- Vital Signs convenience view
CREATE VIEW vw_vital_signs AS
SELECT
    o.id,
    o.patient_id,
    o.encounter_id,
    o.code_value AS loinc_code,
    o.code_display AS vital_name,
    o.value_quantity,
    o.value_unit,
    o.effective_datetime,
    o.interpretation_code,
    o.status
FROM observation o
WHERE o.category_code = 'vital-signs'
  AND o.status IN ('final', 'amended', 'corrected');
