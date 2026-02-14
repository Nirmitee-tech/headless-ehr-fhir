-- ============================================================================
-- T3 OBSTETRICS TABLES MIGRATION
-- Tables: pregnancy, prenatal_visit, labor_record, labor_cervical_exam,
--         fetal_monitoring, delivery_record, newborn_record, postpartum_record
-- ============================================================================

-- ============================================================================
-- 1. PREGNANCY
-- ============================================================================

CREATE TABLE pregnancy (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id              UUID NOT NULL REFERENCES patient(id),
    status                  VARCHAR(30) NOT NULL DEFAULT 'active',
    onset_date              TIMESTAMPTZ,
    estimated_due_date      TIMESTAMPTZ,
    last_menstrual_period   TIMESTAMPTZ,
    conception_method       VARCHAR(50),
    gravida                 INTEGER,
    para                    INTEGER,
    multiple_gestation      BOOLEAN DEFAULT FALSE,
    number_of_fetuses       INTEGER DEFAULT 1,
    risk_level              VARCHAR(20),
    risk_factors            TEXT,
    blood_type              VARCHAR(10),
    rh_factor               VARCHAR(10),
    pre_pregnancy_weight    DECIMAL(6,2),
    pre_pregnancy_bmi       DECIMAL(5,2),
    primary_provider_id     UUID REFERENCES practitioner(id),
    managing_organization_id UUID REFERENCES organization(id),
    note                    TEXT,
    outcome_date            TIMESTAMPTZ,
    outcome_summary         TEXT,
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 2. PRENATAL VISIT
-- ============================================================================

CREATE TABLE prenatal_visit (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pregnancy_id            UUID NOT NULL REFERENCES pregnancy(id) ON DELETE CASCADE,
    encounter_id            UUID REFERENCES encounter(id),
    visit_date              TIMESTAMPTZ NOT NULL,
    gestational_age_weeks   INTEGER,
    gestational_age_days    INTEGER,
    weight                  DECIMAL(6,2),
    blood_pressure_systolic INTEGER,
    blood_pressure_diastolic INTEGER,
    fundal_height           DECIMAL(5,2),
    fetal_heart_rate        INTEGER,
    fetal_presentation      VARCHAR(30),
    fetal_movement          VARCHAR(30),
    urine_protein           VARCHAR(20),
    urine_glucose           VARCHAR(20),
    edema                   VARCHAR(20),
    cervical_dilation       DECIMAL(4,1),
    cervical_effacement     INTEGER,
    group_b_strep_status    VARCHAR(20),
    provider_id             UUID REFERENCES practitioner(id),
    note                    TEXT,
    next_visit_date         TIMESTAMPTZ,
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 3. LABOR RECORD
-- ============================================================================

CREATE TABLE labor_record (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pregnancy_id            UUID NOT NULL REFERENCES pregnancy(id),
    encounter_id            UUID REFERENCES encounter(id),
    admission_datetime      TIMESTAMPTZ,
    labor_onset_datetime    TIMESTAMPTZ,
    labor_onset_type        VARCHAR(30),
    membrane_rupture_datetime TIMESTAMPTZ,
    membrane_rupture_type   VARCHAR(30),
    amniotic_fluid_color    VARCHAR(30),
    amniotic_fluid_volume   VARCHAR(30),
    induction_method        VARCHAR(100),
    induction_reason        VARCHAR(255),
    augmentation_method     VARCHAR(100),
    anesthesia_type         VARCHAR(50),
    anesthesia_start        TIMESTAMPTZ,
    status                  VARCHAR(30) NOT NULL DEFAULT 'active',
    attending_provider_id   UUID REFERENCES practitioner(id),
    note                    TEXT,
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 4. LABOR CERVICAL EXAM
-- ============================================================================

CREATE TABLE labor_cervical_exam (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    labor_record_id         UUID NOT NULL REFERENCES labor_record(id) ON DELETE CASCADE,
    exam_datetime           TIMESTAMPTZ NOT NULL,
    dilation_cm             DECIMAL(4,1),
    effacement_pct          INTEGER,
    station                 VARCHAR(10),
    fetal_position          VARCHAR(20),
    membrane_status         VARCHAR(20),
    examiner_id             UUID REFERENCES practitioner(id),
    note                    TEXT
);

-- ============================================================================
-- 5. FETAL MONITORING
-- ============================================================================

CREATE TABLE fetal_monitoring (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    labor_record_id         UUID NOT NULL REFERENCES labor_record(id) ON DELETE CASCADE,
    monitoring_datetime     TIMESTAMPTZ NOT NULL,
    monitoring_type         VARCHAR(30),
    fetal_heart_rate        INTEGER,
    baseline_rate           INTEGER,
    variability             VARCHAR(20),
    accelerations           VARCHAR(30),
    decelerations           VARCHAR(30),
    deceleration_type       VARCHAR(30),
    contraction_frequency   VARCHAR(30),
    contraction_duration    VARCHAR(30),
    contraction_intensity   VARCHAR(20),
    uterine_resting_tone    VARCHAR(20),
    mvus                    INTEGER,
    interpretation          VARCHAR(50),
    category                VARCHAR(20),
    recorder_id             UUID REFERENCES practitioner(id),
    note                    TEXT
);

-- ============================================================================
-- 6. DELIVERY RECORD
-- ============================================================================

CREATE TABLE delivery_record (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pregnancy_id            UUID NOT NULL REFERENCES pregnancy(id),
    labor_record_id         UUID REFERENCES labor_record(id),
    patient_id              UUID NOT NULL REFERENCES patient(id),
    delivery_datetime       TIMESTAMPTZ NOT NULL,
    delivery_method         VARCHAR(50) NOT NULL,
    delivery_type           VARCHAR(30),
    delivering_provider_id  UUID NOT NULL REFERENCES practitioner(id),
    assistant_provider_id   UUID REFERENCES practitioner(id),
    delivery_location_id    UUID REFERENCES location(id),
    birth_order             INTEGER DEFAULT 1,
    placenta_delivery       VARCHAR(30),
    placenta_datetime       TIMESTAMPTZ,
    placenta_intact         BOOLEAN,
    cord_vessels            INTEGER,
    cord_blood_collected    BOOLEAN DEFAULT FALSE,
    episiotomy              BOOLEAN DEFAULT FALSE,
    episiotomy_type         VARCHAR(30),
    laceration_degree       VARCHAR(20),
    repair_method           VARCHAR(50),
    blood_loss_ml           INTEGER,
    complications           TEXT,
    note                    TEXT,
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 7. NEWBORN RECORD
-- ============================================================================

CREATE TABLE newborn_record (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    delivery_id             UUID NOT NULL REFERENCES delivery_record(id),
    patient_id              UUID REFERENCES patient(id),
    birth_datetime          TIMESTAMPTZ NOT NULL,
    sex                     VARCHAR(10),
    birth_weight_grams      INTEGER,
    birth_length_cm         DECIMAL(5,2),
    head_circumference_cm   DECIMAL(5,2),
    apgar_1min              INTEGER,
    apgar_5min              INTEGER,
    apgar_10min             INTEGER,
    resuscitation_type      VARCHAR(50),
    gestational_age_weeks   INTEGER,
    gestational_age_days    INTEGER,
    birth_status            VARCHAR(30),
    nicu_admission          BOOLEAN DEFAULT FALSE,
    nicu_reason             VARCHAR(255),
    vitamin_k_given         BOOLEAN DEFAULT FALSE,
    eye_prophylaxis_given   BOOLEAN DEFAULT FALSE,
    hepatitis_b_given       BOOLEAN DEFAULT FALSE,
    newborn_screening       VARCHAR(30),
    feeding_method          VARCHAR(30),
    note                    TEXT,
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 8. POSTPARTUM RECORD
-- ============================================================================

CREATE TABLE postpartum_record (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pregnancy_id            UUID NOT NULL REFERENCES pregnancy(id),
    patient_id              UUID NOT NULL REFERENCES patient(id),
    encounter_id            UUID REFERENCES encounter(id),
    visit_date              TIMESTAMPTZ NOT NULL,
    days_postpartum         INTEGER,
    weeks_postpartum        INTEGER,
    uterine_involution      VARCHAR(30),
    lochia_type             VARCHAR(20),
    lochia_amount           VARCHAR(20),
    perineum_status         VARCHAR(30),
    incision_status         VARCHAR(30),
    breast_status           VARCHAR(30),
    breastfeeding_status    VARCHAR(30),
    contraception_plan      VARCHAR(100),
    mood_screening_score    INTEGER,
    mood_screening_tool     VARCHAR(50),
    depression_risk         VARCHAR(20),
    blood_pressure_systolic INTEGER,
    blood_pressure_diastolic INTEGER,
    weight                  DECIMAL(6,2),
    provider_id             UUID REFERENCES practitioner(id),
    note                    TEXT,
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- INDEXES
-- ============================================================================

-- Pregnancy indexes
CREATE INDEX idx_pregnancy_patient ON pregnancy(patient_id);
CREATE INDEX idx_pregnancy_status ON pregnancy(status);
CREATE INDEX idx_pregnancy_due_date ON pregnancy(estimated_due_date) WHERE estimated_due_date IS NOT NULL;

-- Prenatal visit indexes
CREATE INDEX idx_prenatal_visit_pregnancy ON prenatal_visit(pregnancy_id);
CREATE INDEX idx_prenatal_visit_date ON prenatal_visit(visit_date DESC);

-- Labor record indexes
CREATE INDEX idx_labor_record_pregnancy ON labor_record(pregnancy_id);
CREATE INDEX idx_labor_record_status ON labor_record(status);

-- Labor cervical exam indexes
CREATE INDEX idx_cervical_exam_labor ON labor_cervical_exam(labor_record_id);
CREATE INDEX idx_cervical_exam_datetime ON labor_cervical_exam(exam_datetime DESC);

-- Fetal monitoring indexes
CREATE INDEX idx_fetal_monitoring_labor ON fetal_monitoring(labor_record_id);
CREATE INDEX idx_fetal_monitoring_datetime ON fetal_monitoring(monitoring_datetime DESC);

-- Delivery record indexes
CREATE INDEX idx_delivery_pregnancy ON delivery_record(pregnancy_id);
CREATE INDEX idx_delivery_patient ON delivery_record(patient_id);
CREATE INDEX idx_delivery_datetime ON delivery_record(delivery_datetime DESC);

-- Newborn record indexes
CREATE INDEX idx_newborn_delivery ON newborn_record(delivery_id);
CREATE INDEX idx_newborn_patient ON newborn_record(patient_id) WHERE patient_id IS NOT NULL;

-- Postpartum record indexes
CREATE INDEX idx_postpartum_pregnancy ON postpartum_record(pregnancy_id);
CREATE INDEX idx_postpartum_patient ON postpartum_record(patient_id);
CREATE INDEX idx_postpartum_date ON postpartum_record(visit_date DESC);
