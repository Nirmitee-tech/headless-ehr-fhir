-- ============================================================================
-- T4 BEHAVIORAL HEALTH TABLES MIGRATION
-- Tables: psychiatric_assessment, safety_plan, legal_hold,
--         seclusion_restraint_event, group_therapy_session,
--         group_therapy_attendance
-- ============================================================================

-- ============================================================================
-- 1. PSYCHIATRIC ASSESSMENT
-- ============================================================================

CREATE TABLE psychiatric_assessment (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    encounter_id        UUID NOT NULL REFERENCES encounter(id),
    assessor_id         UUID NOT NULL REFERENCES practitioner(id),
    assessment_date     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    chief_complaint     TEXT,
    history_present_illness TEXT,
    psychiatric_history TEXT,
    substance_use_history TEXT,
    family_psych_history TEXT,
    mental_status_exam  TEXT,
    appearance          VARCHAR(255),
    behavior            VARCHAR(255),
    speech              VARCHAR(255),
    mood                VARCHAR(255),
    affect              VARCHAR(255),
    thought_process     VARCHAR(255),
    thought_content     VARCHAR(255),
    perceptions         VARCHAR(255),
    cognition           VARCHAR(255),
    insight             VARCHAR(255),
    judgment            VARCHAR(255),
    risk_assessment     TEXT,
    suicide_risk_level  VARCHAR(30),
    homicide_risk_level VARCHAR(30),
    diagnosis_code      VARCHAR(30),
    diagnosis_display   VARCHAR(500),
    diagnosis_system    VARCHAR(255),
    formulation         TEXT,
    treatment_plan      TEXT,
    disposition         VARCHAR(100),
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 2. SAFETY PLAN
-- ============================================================================

CREATE TABLE safety_plan (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    created_by_id       UUID NOT NULL REFERENCES practitioner(id),
    status              VARCHAR(30) NOT NULL DEFAULT 'active',
    plan_date           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    warning_signs       TEXT,
    coping_strategies   TEXT,
    social_distractions TEXT,
    people_to_contact   TEXT,
    professionals_to_contact TEXT,
    emergency_contacts  TEXT,
    means_restriction   TEXT,
    reasons_for_living  TEXT,
    patient_signature   BOOLEAN DEFAULT FALSE,
    provider_signature  BOOLEAN DEFAULT FALSE,
    review_date         TIMESTAMPTZ,
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 3. LEGAL HOLD
-- ============================================================================

CREATE TABLE legal_hold (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    encounter_id        UUID REFERENCES encounter(id),
    initiated_by_id     UUID NOT NULL REFERENCES practitioner(id),
    status              VARCHAR(30) NOT NULL DEFAULT 'active',
    hold_type           VARCHAR(50) NOT NULL,
    authority_statute   VARCHAR(255),
    start_datetime      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    end_datetime        TIMESTAMPTZ,
    duration_hours      INTEGER,
    reason              TEXT NOT NULL,
    criteria_met        TEXT,
    certifying_physician_id UUID REFERENCES practitioner(id),
    certification_datetime  TIMESTAMPTZ,
    court_hearing_date  TIMESTAMPTZ,
    court_order_number  VARCHAR(100),
    legal_counsel_notified BOOLEAN DEFAULT FALSE,
    patient_rights_given   BOOLEAN DEFAULT FALSE,
    release_reason      VARCHAR(255),
    release_authorized_by_id UUID REFERENCES practitioner(id),
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 4. SECLUSION / RESTRAINT EVENT
-- ============================================================================

CREATE TABLE seclusion_restraint_event (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    encounter_id        UUID REFERENCES encounter(id),
    ordered_by_id       UUID NOT NULL REFERENCES practitioner(id),
    event_type          VARCHAR(30) NOT NULL,
    restraint_type      VARCHAR(50),
    start_datetime      TIMESTAMPTZ NOT NULL,
    end_datetime        TIMESTAMPTZ,
    reason              TEXT NOT NULL,
    behavior_description TEXT,
    alternatives_attempted TEXT,
    monitoring_frequency_min INTEGER,
    last_monitoring_check TIMESTAMPTZ,
    patient_condition_during TEXT,
    injuries_noted      TEXT,
    nutrition_offered   BOOLEAN DEFAULT FALSE,
    toileting_offered   BOOLEAN DEFAULT FALSE,
    discontinued_by_id  UUID REFERENCES practitioner(id),
    discontinuation_reason VARCHAR(255),
    debrief_completed   BOOLEAN DEFAULT FALSE,
    debrief_notes       TEXT,
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 5. GROUP THERAPY SESSION
-- ============================================================================

CREATE TABLE group_therapy_session (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_name        VARCHAR(255) NOT NULL,
    session_type        VARCHAR(100),
    facilitator_id      UUID NOT NULL REFERENCES practitioner(id),
    co_facilitator_id   UUID REFERENCES practitioner(id),
    status              VARCHAR(30) NOT NULL DEFAULT 'scheduled',
    scheduled_datetime  TIMESTAMPTZ NOT NULL,
    actual_start        TIMESTAMPTZ,
    actual_end          TIMESTAMPTZ,
    location            VARCHAR(255),
    max_participants    INTEGER,
    topic               TEXT,
    session_goals       TEXT,
    session_notes       TEXT,
    materials_used      TEXT,
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 6. GROUP THERAPY ATTENDANCE
-- ============================================================================

CREATE TABLE group_therapy_attendance (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id          UUID NOT NULL REFERENCES group_therapy_session(id) ON DELETE CASCADE,
    patient_id          UUID NOT NULL REFERENCES patient(id),
    attendance_status   VARCHAR(30) NOT NULL DEFAULT 'present',
    participation_level VARCHAR(30),
    behavior_notes      TEXT,
    mood_before         VARCHAR(50),
    mood_after          VARCHAR(50),
    note                TEXT,
    UNIQUE(session_id, patient_id)
);

-- ============================================================================
-- 7. INDEXES
-- ============================================================================

-- Psychiatric Assessment indexes
CREATE INDEX idx_psych_assess_patient ON psychiatric_assessment(patient_id);
CREATE INDEX idx_psych_assess_encounter ON psychiatric_assessment(encounter_id);
CREATE INDEX idx_psych_assess_assessor ON psychiatric_assessment(assessor_id);
CREATE INDEX idx_psych_assess_date ON psychiatric_assessment(assessment_date DESC);

-- Safety Plan indexes
CREATE INDEX idx_safety_plan_patient ON safety_plan(patient_id);
CREATE INDEX idx_safety_plan_status ON safety_plan(status);
CREATE INDEX idx_safety_plan_created_by ON safety_plan(created_by_id);

-- Legal Hold indexes
CREATE INDEX idx_legal_hold_patient ON legal_hold(patient_id);
CREATE INDEX idx_legal_hold_status ON legal_hold(status);
CREATE INDEX idx_legal_hold_start ON legal_hold(start_datetime DESC);
CREATE INDEX idx_legal_hold_encounter ON legal_hold(encounter_id) WHERE encounter_id IS NOT NULL;

-- Seclusion/Restraint indexes
CREATE INDEX idx_seclusion_patient ON seclusion_restraint_event(patient_id);
CREATE INDEX idx_seclusion_encounter ON seclusion_restraint_event(encounter_id) WHERE encounter_id IS NOT NULL;
CREATE INDEX idx_seclusion_start ON seclusion_restraint_event(start_datetime DESC);
CREATE INDEX idx_seclusion_type ON seclusion_restraint_event(event_type);

-- Group Therapy indexes
CREATE INDEX idx_group_therapy_facilitator ON group_therapy_session(facilitator_id);
CREATE INDEX idx_group_therapy_status ON group_therapy_session(status);
CREATE INDEX idx_group_therapy_scheduled ON group_therapy_session(scheduled_datetime DESC);
CREATE INDEX idx_group_attendance_session ON group_therapy_attendance(session_id);
CREATE INDEX idx_group_attendance_patient ON group_therapy_attendance(patient_id);
