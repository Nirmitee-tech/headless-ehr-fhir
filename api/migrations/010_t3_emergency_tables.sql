-- ============================================================================
-- T3 EMERGENCY TABLES MIGRATION
-- Tables: triage_record, ed_tracking, ed_status_history, trauma_activation
-- ============================================================================

-- ============================================================================
-- 1. TRIAGE RECORD
-- ============================================================================

CREATE TABLE triage_record (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    encounter_id        UUID NOT NULL REFERENCES encounter(id),
    triage_nurse_id     UUID NOT NULL REFERENCES practitioner(id),
    arrival_time        TIMESTAMPTZ,
    triage_time         TIMESTAMPTZ DEFAULT NOW(),
    chief_complaint     TEXT NOT NULL,
    acuity_level        INTEGER,
    acuity_system       VARCHAR(50),
    pain_scale          INTEGER,
    arrival_mode        VARCHAR(30),
    heart_rate          INTEGER,
    blood_pressure_sys  INTEGER,
    blood_pressure_dia  INTEGER,
    temperature         DECIMAL(5,2),
    respiratory_rate    INTEGER,
    oxygen_saturation   INTEGER,
    glasgow_coma_score  INTEGER,
    injury_description  TEXT,
    allergy_note        TEXT,
    medication_note     TEXT,
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 2. ED TRACKING
-- ============================================================================

CREATE TABLE ed_tracking (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    encounter_id        UUID NOT NULL REFERENCES encounter(id),
    triage_record_id    UUID REFERENCES triage_record(id),
    current_status      VARCHAR(30) NOT NULL DEFAULT 'waiting',
    bed_assignment      VARCHAR(30),
    attending_id        UUID REFERENCES practitioner(id),
    nurse_id            UUID REFERENCES practitioner(id),
    arrival_time        TIMESTAMPTZ,
    discharge_time      TIMESTAMPTZ,
    disposition         VARCHAR(50),
    disposition_dest    VARCHAR(255),
    length_of_stay_mins INTEGER,
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 3. ED STATUS HISTORY
-- ============================================================================

CREATE TABLE ed_status_history (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ed_tracking_id      UUID NOT NULL REFERENCES ed_tracking(id) ON DELETE CASCADE,
    status              VARCHAR(30) NOT NULL,
    changed_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    changed_by          UUID REFERENCES practitioner(id),
    note                TEXT
);

-- ============================================================================
-- 4. TRAUMA ACTIVATION
-- ============================================================================

CREATE TABLE trauma_activation (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    encounter_id        UUID REFERENCES encounter(id),
    ed_tracking_id      UUID REFERENCES ed_tracking(id),
    activation_level    VARCHAR(30) NOT NULL,
    activation_time     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deactivation_time   TIMESTAMPTZ,
    mechanism_of_injury TEXT,
    activated_by        UUID REFERENCES practitioner(id),
    team_lead_id        UUID REFERENCES practitioner(id),
    outcome             VARCHAR(100),
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 5. INDEXES for T3 Emergency tables
-- ============================================================================

-- Triage Record indexes
CREATE INDEX idx_triage_patient ON triage_record(patient_id);
CREATE INDEX idx_triage_encounter ON triage_record(encounter_id);
CREATE INDEX idx_triage_nurse ON triage_record(triage_nurse_id);
CREATE INDEX idx_triage_acuity ON triage_record(acuity_level) WHERE acuity_level IS NOT NULL;
CREATE INDEX idx_triage_time ON triage_record(triage_time DESC) WHERE triage_time IS NOT NULL;

-- ED Tracking indexes
CREATE INDEX idx_ed_track_patient ON ed_tracking(patient_id);
CREATE INDEX idx_ed_track_encounter ON ed_tracking(encounter_id);
CREATE INDEX idx_ed_track_status ON ed_tracking(current_status);
CREATE INDEX idx_ed_track_attending ON ed_tracking(attending_id) WHERE attending_id IS NOT NULL;
CREATE INDEX idx_ed_track_triage ON ed_tracking(triage_record_id) WHERE triage_record_id IS NOT NULL;

-- ED Status History indexes
CREATE INDEX idx_ed_status_tracking ON ed_status_history(ed_tracking_id);
CREATE INDEX idx_ed_status_time ON ed_status_history(changed_at DESC);

-- Trauma Activation indexes
CREATE INDEX idx_trauma_patient ON trauma_activation(patient_id);
CREATE INDEX idx_trauma_level ON trauma_activation(activation_level);
CREATE INDEX idx_trauma_time ON trauma_activation(activation_time DESC);
CREATE INDEX idx_trauma_ed_track ON trauma_activation(ed_tracking_id) WHERE ed_tracking_id IS NOT NULL;
