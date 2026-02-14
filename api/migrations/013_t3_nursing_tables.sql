-- ============================================================================
-- T3 NURSING TABLES MIGRATION
-- Tables: flowsheet_template, flowsheet_row, flowsheet_entry,
--         nursing_assessment, fall_risk_assessment, skin_assessment,
--         pain_assessment, lines_drains_airways, restraint_record,
--         intake_output_record
-- ============================================================================

-- ============================================================================
-- 1. FLOWSHEET TEMPLATE
-- ============================================================================

CREATE TABLE flowsheet_template (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    category        VARCHAR(100),
    is_active       BOOLEAN DEFAULT TRUE,
    created_by      UUID REFERENCES practitioner(id),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 2. FLOWSHEET ROW
-- ============================================================================

CREATE TABLE flowsheet_row (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id     UUID NOT NULL REFERENCES flowsheet_template(id) ON DELETE CASCADE,
    label           VARCHAR(255) NOT NULL,
    data_type       VARCHAR(50) NOT NULL,
    unit            VARCHAR(50),
    allowed_values  TEXT[],
    sort_order      INTEGER DEFAULT 0,
    is_required     BOOLEAN DEFAULT FALSE
);

-- ============================================================================
-- 3. FLOWSHEET ENTRY
-- ============================================================================

CREATE TABLE flowsheet_entry (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id     UUID NOT NULL REFERENCES flowsheet_template(id),
    row_id          UUID NOT NULL REFERENCES flowsheet_row(id),
    patient_id      UUID NOT NULL REFERENCES patient(id),
    encounter_id    UUID NOT NULL REFERENCES encounter(id),
    value_text      TEXT,
    value_numeric   DECIMAL(12,4),
    recorded_at     TIMESTAMPTZ DEFAULT NOW(),
    recorded_by_id  UUID NOT NULL REFERENCES practitioner(id),
    note            TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 4. NURSING ASSESSMENT
-- ============================================================================

CREATE TABLE nursing_assessment (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id      UUID NOT NULL REFERENCES patient(id),
    encounter_id    UUID NOT NULL REFERENCES encounter(id),
    nurse_id        UUID NOT NULL REFERENCES practitioner(id),
    assessment_type VARCHAR(100) NOT NULL,
    assessment_data TEXT,
    status          VARCHAR(30) NOT NULL DEFAULT 'in-progress',
    completed_at    TIMESTAMPTZ,
    note            TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 5. FALL RISK ASSESSMENT
-- ============================================================================

CREATE TABLE fall_risk_assessment (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id      UUID NOT NULL REFERENCES patient(id),
    encounter_id    UUID REFERENCES encounter(id),
    assessed_by_id  UUID NOT NULL REFERENCES practitioner(id),
    tool_used       VARCHAR(100),
    total_score     INTEGER,
    risk_level      VARCHAR(30),
    history_of_falls BOOLEAN,
    medications     BOOLEAN,
    gait_balance    VARCHAR(100),
    mental_status   VARCHAR(100),
    interventions   TEXT,
    note            TEXT,
    assessed_at     TIMESTAMPTZ DEFAULT NOW(),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 6. SKIN ASSESSMENT
-- ============================================================================

CREATE TABLE skin_assessment (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id      UUID NOT NULL REFERENCES patient(id),
    encounter_id    UUID REFERENCES encounter(id),
    assessed_by_id  UUID NOT NULL REFERENCES practitioner(id),
    tool_used       VARCHAR(100),
    total_score     INTEGER,
    risk_level      VARCHAR(30),
    skin_integrity  VARCHAR(100),
    moisture_level  VARCHAR(50),
    mobility        VARCHAR(50),
    nutrition       VARCHAR(50),
    wound_present   BOOLEAN,
    wound_location  VARCHAR(255),
    wound_stage     VARCHAR(50),
    interventions   TEXT,
    note            TEXT,
    assessed_at     TIMESTAMPTZ DEFAULT NOW(),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 7. PAIN ASSESSMENT
-- ============================================================================

CREATE TABLE pain_assessment (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id      UUID NOT NULL REFERENCES patient(id),
    encounter_id    UUID REFERENCES encounter(id),
    assessed_by_id  UUID NOT NULL REFERENCES practitioner(id),
    tool_used       VARCHAR(100),
    pain_score      INTEGER,
    pain_location   VARCHAR(255),
    pain_character  VARCHAR(100),
    pain_duration   VARCHAR(100),
    pain_radiation  VARCHAR(255),
    aggravating     TEXT,
    alleviating     TEXT,
    interventions   TEXT,
    reassess_score  INTEGER,
    note            TEXT,
    assessed_at     TIMESTAMPTZ DEFAULT NOW(),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 8. LINES, DRAINS & AIRWAYS
-- ============================================================================

CREATE TABLE lines_drains_airways (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id      UUID NOT NULL REFERENCES patient(id),
    encounter_id    UUID NOT NULL REFERENCES encounter(id),
    type            VARCHAR(100) NOT NULL,
    description     TEXT,
    site            VARCHAR(255),
    size            VARCHAR(50),
    inserted_at     TIMESTAMPTZ,
    inserted_by_id  UUID REFERENCES practitioner(id),
    removed_at      TIMESTAMPTZ,
    removed_by_id   UUID REFERENCES practitioner(id),
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    device_id       UUID,
    note            TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 9. RESTRAINT RECORD
-- ============================================================================

CREATE TABLE restraint_record (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    encounter_id        UUID REFERENCES encounter(id),
    restraint_type      VARCHAR(100) NOT NULL,
    reason              TEXT,
    body_site           VARCHAR(255),
    applied_at          TIMESTAMPTZ NOT NULL,
    applied_by_id       UUID NOT NULL REFERENCES practitioner(id),
    removed_at          TIMESTAMPTZ,
    removed_by_id       UUID REFERENCES practitioner(id),
    order_id            UUID,
    last_assessed_at    TIMESTAMPTZ,
    last_assessed_by_id UUID REFERENCES practitioner(id),
    skin_condition      VARCHAR(100),
    circulation         VARCHAR(100),
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 10. INTAKE/OUTPUT RECORD
-- ============================================================================

CREATE TABLE intake_output_record (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id      UUID NOT NULL REFERENCES patient(id),
    encounter_id    UUID NOT NULL REFERENCES encounter(id),
    category        VARCHAR(30) NOT NULL,
    type            VARCHAR(100),
    volume          DECIMAL(10,2),
    unit            VARCHAR(20),
    route           VARCHAR(50),
    recorded_at     TIMESTAMPTZ DEFAULT NOW(),
    recorded_by_id  UUID NOT NULL REFERENCES practitioner(id),
    note            TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 11. INDEXES for T3 nursing tables
-- ============================================================================

-- Flowsheet template indexes
CREATE INDEX idx_flowsheet_template_active ON flowsheet_template(is_active);

-- Flowsheet row indexes
CREATE INDEX idx_flowsheet_row_template ON flowsheet_row(template_id);

-- Flowsheet entry indexes
CREATE INDEX idx_flowsheet_entry_patient ON flowsheet_entry(patient_id);
CREATE INDEX idx_flowsheet_entry_encounter ON flowsheet_entry(encounter_id);
CREATE INDEX idx_flowsheet_entry_template ON flowsheet_entry(template_id);
CREATE INDEX idx_flowsheet_entry_row ON flowsheet_entry(row_id);
CREATE INDEX idx_flowsheet_entry_recorded ON flowsheet_entry(recorded_at DESC);

-- Nursing assessment indexes
CREATE INDEX idx_nursing_assessment_patient ON nursing_assessment(patient_id);
CREATE INDEX idx_nursing_assessment_encounter ON nursing_assessment(encounter_id);
CREATE INDEX idx_nursing_assessment_nurse ON nursing_assessment(nurse_id);
CREATE INDEX idx_nursing_assessment_type ON nursing_assessment(assessment_type);

-- Fall risk assessment indexes
CREATE INDEX idx_fall_risk_patient ON fall_risk_assessment(patient_id);
CREATE INDEX idx_fall_risk_assessed_at ON fall_risk_assessment(assessed_at DESC);

-- Skin assessment indexes
CREATE INDEX idx_skin_assessment_patient ON skin_assessment(patient_id);
CREATE INDEX idx_skin_assessment_assessed_at ON skin_assessment(assessed_at DESC);

-- Pain assessment indexes
CREATE INDEX idx_pain_assessment_patient ON pain_assessment(patient_id);
CREATE INDEX idx_pain_assessment_assessed_at ON pain_assessment(assessed_at DESC);

-- Lines/drains/airways indexes
CREATE INDEX idx_lines_drains_patient ON lines_drains_airways(patient_id);
CREATE INDEX idx_lines_drains_encounter ON lines_drains_airways(encounter_id);
CREATE INDEX idx_lines_drains_status ON lines_drains_airways(status);

-- Restraint indexes
CREATE INDEX idx_restraint_patient ON restraint_record(patient_id);
CREATE INDEX idx_restraint_applied ON restraint_record(applied_at DESC);

-- Intake/output indexes
CREATE INDEX idx_intake_output_patient ON intake_output_record(patient_id);
CREATE INDEX idx_intake_output_encounter ON intake_output_record(encounter_id);
CREATE INDEX idx_intake_output_category ON intake_output_record(category);
CREATE INDEX idx_intake_output_recorded ON intake_output_record(recorded_at DESC);
