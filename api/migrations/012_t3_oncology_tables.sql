-- ============================================================================
-- T3 ONCOLOGY TABLES MIGRATION
-- Tables: cancer_diagnosis, treatment_protocol, treatment_protocol_drug,
--         chemotherapy_cycle, chemotherapy_administration, radiation_therapy,
--         radiation_therapy_session, tumor_marker, tumor_board_review
-- ============================================================================

-- ============================================================================
-- 1. CANCER DIAGNOSIS
-- ============================================================================

CREATE TABLE cancer_diagnosis (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id              UUID NOT NULL REFERENCES patient(id),
    condition_id            UUID REFERENCES condition(id),
    diagnosis_date          TIMESTAMPTZ NOT NULL,
    cancer_type             VARCHAR(100),
    cancer_site             VARCHAR(100),
    histology_code          VARCHAR(30),
    histology_display       VARCHAR(255),
    morphology_code         VARCHAR(30),
    morphology_display      VARCHAR(255),
    staging_system          VARCHAR(50),
    stage_group             VARCHAR(20),
    t_stage                 VARCHAR(20),
    n_stage                 VARCHAR(20),
    m_stage                 VARCHAR(20),
    grade                   VARCHAR(20),
    laterality              VARCHAR(20),
    current_status          VARCHAR(30) NOT NULL DEFAULT 'active-treatment',
    diagnosing_provider_id  UUID REFERENCES practitioner(id),
    managing_provider_id    UUID REFERENCES practitioner(id),
    icd10_code              VARCHAR(20),
    icd10_display           VARCHAR(255),
    icdo3_topography        VARCHAR(20),
    icdo3_morphology        VARCHAR(20),
    note                    TEXT,
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 2. TREATMENT PROTOCOL
-- ============================================================================

CREATE TABLE treatment_protocol (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cancer_diagnosis_id     UUID NOT NULL REFERENCES cancer_diagnosis(id),
    protocol_name           VARCHAR(255) NOT NULL,
    protocol_code           VARCHAR(50),
    protocol_type           VARCHAR(50),
    intent                  VARCHAR(30),
    number_of_cycles        INTEGER,
    cycle_length_days       INTEGER,
    start_date              TIMESTAMPTZ,
    end_date                TIMESTAMPTZ,
    status                  VARCHAR(30) NOT NULL DEFAULT 'planned',
    prescribing_provider_id UUID REFERENCES practitioner(id),
    clinical_trial_id       VARCHAR(100),
    note                    TEXT,
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 3. TREATMENT PROTOCOL DRUG
-- ============================================================================

CREATE TABLE treatment_protocol_drug (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    protocol_id             UUID NOT NULL REFERENCES treatment_protocol(id) ON DELETE CASCADE,
    drug_name               VARCHAR(255) NOT NULL,
    drug_code               VARCHAR(50),
    drug_code_system        VARCHAR(100),
    route                   VARCHAR(30),
    dose_value              DECIMAL(10,3),
    dose_unit               VARCHAR(30),
    dose_calculation_method VARCHAR(50),
    frequency               VARCHAR(50),
    administration_day      VARCHAR(50),
    infusion_duration_min   INTEGER,
    premedication           TEXT,
    sequence_order          INTEGER,
    note                    TEXT
);

-- ============================================================================
-- 4. CHEMOTHERAPY CYCLE
-- ============================================================================

CREATE TABLE chemotherapy_cycle (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    protocol_id             UUID NOT NULL REFERENCES treatment_protocol(id),
    cycle_number            INTEGER NOT NULL,
    planned_start_date      TIMESTAMPTZ,
    actual_start_date       TIMESTAMPTZ,
    actual_end_date         TIMESTAMPTZ,
    status                  VARCHAR(30) NOT NULL DEFAULT 'planned',
    dose_reduction_pct      DECIMAL(5,2),
    dose_reduction_reason   VARCHAR(255),
    delay_days              INTEGER,
    delay_reason            VARCHAR(255),
    bsa_m2                  DECIMAL(5,3),
    weight_kg               DECIMAL(6,2),
    height_cm               DECIMAL(5,1),
    creatinine_clearance    DECIMAL(6,2),
    provider_id             UUID REFERENCES practitioner(id),
    note                    TEXT,
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 5. CHEMOTHERAPY ADMINISTRATION
-- ============================================================================

CREATE TABLE chemotherapy_administration (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cycle_id                UUID NOT NULL REFERENCES chemotherapy_cycle(id) ON DELETE CASCADE,
    protocol_drug_id        UUID REFERENCES treatment_protocol_drug(id),
    drug_name               VARCHAR(255) NOT NULL,
    administration_datetime TIMESTAMPTZ NOT NULL,
    dose_given              DECIMAL(10,3),
    dose_unit               VARCHAR(30),
    route                   VARCHAR(30),
    infusion_duration_min   INTEGER,
    infusion_rate           VARCHAR(50),
    site                    VARCHAR(50),
    sequence_number         INTEGER,
    reaction_type           VARCHAR(50),
    reaction_severity       VARCHAR(20),
    reaction_action         VARCHAR(100),
    administering_nurse_id  UUID REFERENCES practitioner(id),
    supervising_provider_id UUID REFERENCES practitioner(id),
    note                    TEXT
);

-- ============================================================================
-- 6. RADIATION THERAPY
-- ============================================================================

CREATE TABLE radiation_therapy (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cancer_diagnosis_id     UUID NOT NULL REFERENCES cancer_diagnosis(id),
    therapy_type            VARCHAR(50),
    modality                VARCHAR(50),
    technique               VARCHAR(50),
    target_site             VARCHAR(100),
    laterality              VARCHAR(20),
    total_dose_cgy          DECIMAL(10,2),
    dose_per_fraction_cgy   DECIMAL(8,2),
    planned_fractions       INTEGER,
    completed_fractions     INTEGER DEFAULT 0,
    start_date              TIMESTAMPTZ,
    end_date                TIMESTAMPTZ,
    status                  VARCHAR(30) NOT NULL DEFAULT 'planned',
    prescribing_provider_id UUID REFERENCES practitioner(id),
    treating_facility_id    UUID REFERENCES organization(id),
    energy_type             VARCHAR(30),
    energy_value            VARCHAR(30),
    treatment_volume_cc     DECIMAL(8,2),
    note                    TEXT,
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 7. RADIATION THERAPY SESSION
-- ============================================================================

CREATE TABLE radiation_therapy_session (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    radiation_therapy_id    UUID NOT NULL REFERENCES radiation_therapy(id) ON DELETE CASCADE,
    session_number          INTEGER NOT NULL,
    session_date            TIMESTAMPTZ NOT NULL,
    dose_delivered_cgy      DECIMAL(8,2),
    field_name              VARCHAR(50),
    setup_verified          BOOLEAN DEFAULT FALSE,
    imaging_type            VARCHAR(30),
    skin_reaction_grade     INTEGER,
    fatigue_grade           INTEGER,
    other_toxicity          VARCHAR(255),
    toxicity_grade          INTEGER,
    machine_id              VARCHAR(50),
    therapist_id            UUID REFERENCES practitioner(id),
    physicist_id            UUID REFERENCES practitioner(id),
    note                    TEXT
);

-- ============================================================================
-- 8. TUMOR MARKER
-- ============================================================================

CREATE TABLE tumor_marker (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cancer_diagnosis_id     UUID REFERENCES cancer_diagnosis(id),
    patient_id              UUID NOT NULL REFERENCES patient(id),
    marker_name             VARCHAR(100) NOT NULL,
    marker_code             VARCHAR(30),
    marker_code_system      VARCHAR(100),
    value_quantity          DECIMAL(12,4),
    value_unit              VARCHAR(30),
    value_string            VARCHAR(255),
    value_interpretation    VARCHAR(30),
    reference_range_low     DECIMAL(12,4),
    reference_range_high    DECIMAL(12,4),
    reference_range_text    VARCHAR(100),
    specimen_type           VARCHAR(50),
    collection_datetime     TIMESTAMPTZ,
    result_datetime         TIMESTAMPTZ,
    performing_lab          VARCHAR(255),
    ordering_provider_id    UUID REFERENCES practitioner(id),
    note                    TEXT,
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 9. TUMOR BOARD REVIEW
-- ============================================================================

CREATE TABLE tumor_board_review (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cancer_diagnosis_id     UUID NOT NULL REFERENCES cancer_diagnosis(id),
    patient_id              UUID NOT NULL REFERENCES patient(id),
    review_date             TIMESTAMPTZ NOT NULL,
    review_type             VARCHAR(50),
    presenting_provider_id  UUID REFERENCES practitioner(id),
    attendees               TEXT,
    clinical_summary        TEXT,
    pathology_summary       TEXT,
    imaging_summary         TEXT,
    discussion              TEXT,
    recommendations         TEXT,
    treatment_decision      VARCHAR(255),
    clinical_trial_discussed BOOLEAN DEFAULT FALSE,
    clinical_trial_id       VARCHAR(100),
    next_review_date        TIMESTAMPTZ,
    note                    TEXT,
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- INDEXES
-- ============================================================================

-- Cancer diagnosis indexes
CREATE INDEX idx_cancer_dx_patient ON cancer_diagnosis(patient_id);
CREATE INDEX idx_cancer_dx_date ON cancer_diagnosis(diagnosis_date DESC);
CREATE INDEX idx_cancer_dx_status ON cancer_diagnosis(current_status);
CREATE INDEX idx_cancer_dx_type ON cancer_diagnosis(cancer_type) WHERE cancer_type IS NOT NULL;

-- Treatment protocol indexes
CREATE INDEX idx_treatment_protocol_dx ON treatment_protocol(cancer_diagnosis_id);
CREATE INDEX idx_treatment_protocol_status ON treatment_protocol(status);

-- Treatment protocol drug indexes
CREATE INDEX idx_protocol_drug_protocol ON treatment_protocol_drug(protocol_id);

-- Chemotherapy cycle indexes
CREATE INDEX idx_chemo_cycle_protocol ON chemotherapy_cycle(protocol_id);
CREATE INDEX idx_chemo_cycle_status ON chemotherapy_cycle(status);
CREATE INDEX idx_chemo_cycle_start ON chemotherapy_cycle(planned_start_date) WHERE planned_start_date IS NOT NULL;

-- Chemotherapy administration indexes
CREATE INDEX idx_chemo_admin_cycle ON chemotherapy_administration(cycle_id);
CREATE INDEX idx_chemo_admin_datetime ON chemotherapy_administration(administration_datetime DESC);

-- Radiation therapy indexes
CREATE INDEX idx_radiation_dx ON radiation_therapy(cancer_diagnosis_id);
CREATE INDEX idx_radiation_status ON radiation_therapy(status);
CREATE INDEX idx_radiation_start ON radiation_therapy(start_date) WHERE start_date IS NOT NULL;

-- Radiation therapy session indexes
CREATE INDEX idx_radiation_session_therapy ON radiation_therapy_session(radiation_therapy_id);
CREATE INDEX idx_radiation_session_date ON radiation_therapy_session(session_date DESC);

-- Tumor marker indexes
CREATE INDEX idx_tumor_marker_patient ON tumor_marker(patient_id);
CREATE INDEX idx_tumor_marker_dx ON tumor_marker(cancer_diagnosis_id) WHERE cancer_diagnosis_id IS NOT NULL;
CREATE INDEX idx_tumor_marker_name ON tumor_marker(marker_name);
CREATE INDEX idx_tumor_marker_collection ON tumor_marker(collection_datetime DESC) WHERE collection_datetime IS NOT NULL;

-- Tumor board review indexes
CREATE INDEX idx_tumor_board_dx ON tumor_board_review(cancer_diagnosis_id);
CREATE INDEX idx_tumor_board_patient ON tumor_board_review(patient_id);
CREATE INDEX idx_tumor_board_date ON tumor_board_review(review_date DESC);
