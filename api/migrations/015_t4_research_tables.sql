-- ============================================================================
-- T4 RESEARCH TABLES MIGRATION
-- Tables: research_study, research_arm, research_enrollment,
--         research_adverse_event, research_protocol_deviation
-- ============================================================================

-- ============================================================================
-- 1. RESEARCH STUDY
-- ============================================================================

CREATE TABLE research_study (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    title               VARCHAR(500) NOT NULL,
    protocol_number     VARCHAR(100) NOT NULL UNIQUE,
    status              VARCHAR(50) NOT NULL DEFAULT 'in-review',
    phase               VARCHAR(30),
    category            VARCHAR(100),
    focus               VARCHAR(255),
    description         TEXT,
    sponsor_name        VARCHAR(255),
    sponsor_contact     VARCHAR(255),
    principal_investigator_id UUID REFERENCES practitioner(id),
    site_name           VARCHAR(255),
    site_contact        VARCHAR(255),
    irb_number          VARCHAR(100),
    irb_approval_date   DATE,
    irb_expiration_date DATE,
    start_date          DATE,
    end_date            DATE,
    enrollment_target   INTEGER,
    enrollment_actual   INTEGER DEFAULT 0,
    primary_endpoint    TEXT,
    secondary_endpoints TEXT,
    inclusion_criteria  TEXT,
    exclusion_criteria  TEXT,
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 2. RESEARCH ARM
-- ============================================================================

CREATE TABLE research_arm (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    study_id            UUID NOT NULL REFERENCES research_study(id) ON DELETE CASCADE,
    name                VARCHAR(255) NOT NULL,
    arm_type            VARCHAR(50),
    description         TEXT,
    target_enrollment   INTEGER,
    actual_enrollment   INTEGER DEFAULT 0
);

-- ============================================================================
-- 3. RESEARCH ENROLLMENT
-- ============================================================================

CREATE TABLE research_enrollment (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    study_id            UUID NOT NULL REFERENCES research_study(id),
    arm_id              UUID REFERENCES research_arm(id),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    consent_id          UUID,
    status              VARCHAR(50) NOT NULL DEFAULT 'pre-screening',
    enrolled_date       DATE,
    screening_date      DATE,
    randomization_date  DATE,
    completion_date     DATE,
    withdrawal_date     DATE,
    withdrawal_reason   VARCHAR(255),
    randomization_number VARCHAR(50),
    subject_number      VARCHAR(50),
    enrolled_by_id      UUID REFERENCES practitioner(id),
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(study_id, patient_id)
);

-- ============================================================================
-- 4. RESEARCH ADVERSE EVENT
-- ============================================================================

CREATE TABLE research_adverse_event (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    enrollment_id       UUID NOT NULL REFERENCES research_enrollment(id),
    event_date          TIMESTAMPTZ NOT NULL,
    reported_date       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reported_by_id      UUID REFERENCES practitioner(id),
    description         TEXT NOT NULL,
    severity            VARCHAR(30),
    seriousness         VARCHAR(30),
    causality           VARCHAR(50),
    expectedness        VARCHAR(30),
    outcome             VARCHAR(50),
    action_taken        VARCHAR(100),
    resolution_date     TIMESTAMPTZ,
    reported_to_irb     BOOLEAN DEFAULT FALSE,
    irb_report_date     TIMESTAMPTZ,
    reported_to_sponsor BOOLEAN DEFAULT FALSE,
    sponsor_report_date TIMESTAMPTZ,
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 5. RESEARCH PROTOCOL DEVIATION
-- ============================================================================

CREATE TABLE research_protocol_deviation (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    enrollment_id       UUID NOT NULL REFERENCES research_enrollment(id),
    deviation_date      TIMESTAMPTZ NOT NULL,
    reported_date       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reported_by_id      UUID REFERENCES practitioner(id),
    category            VARCHAR(100),
    description         TEXT NOT NULL,
    severity            VARCHAR(30),
    corrective_action   TEXT,
    preventive_action   TEXT,
    impact_on_subject   TEXT,
    impact_on_study     TEXT,
    reported_to_irb     BOOLEAN DEFAULT FALSE,
    irb_report_date     TIMESTAMPTZ,
    reported_to_sponsor BOOLEAN DEFAULT FALSE,
    sponsor_report_date TIMESTAMPTZ,
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 6. INDEXES
-- ============================================================================

-- Research Study indexes
CREATE INDEX idx_research_study_fhir ON research_study(fhir_id);
CREATE INDEX idx_research_study_status ON research_study(status);
CREATE INDEX idx_research_study_protocol ON research_study(protocol_number);
CREATE INDEX idx_research_study_pi ON research_study(principal_investigator_id) WHERE principal_investigator_id IS NOT NULL;

-- Research Arm indexes
CREATE INDEX idx_research_arm_study ON research_arm(study_id);

-- Research Enrollment indexes
CREATE INDEX idx_research_enrollment_study ON research_enrollment(study_id);
CREATE INDEX idx_research_enrollment_patient ON research_enrollment(patient_id);
CREATE INDEX idx_research_enrollment_status ON research_enrollment(status);
CREATE INDEX idx_research_enrollment_arm ON research_enrollment(arm_id) WHERE arm_id IS NOT NULL;

-- Adverse Event indexes
CREATE INDEX idx_research_ae_enrollment ON research_adverse_event(enrollment_id);
CREATE INDEX idx_research_ae_date ON research_adverse_event(event_date DESC);
CREATE INDEX idx_research_ae_severity ON research_adverse_event(severity) WHERE severity IS NOT NULL;

-- Protocol Deviation indexes
CREATE INDEX idx_research_pd_enrollment ON research_protocol_deviation(enrollment_id);
CREATE INDEX idx_research_pd_date ON research_protocol_deviation(deviation_date DESC);
CREATE INDEX idx_research_pd_category ON research_protocol_deviation(category) WHERE category IS NOT NULL;
