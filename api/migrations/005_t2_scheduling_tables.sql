-- ============================================================================
-- T2 SCHEDULING TABLES MIGRATION
-- Tables: schedule, slot, appointment, appointment_participant,
--         appointment_response, waitlist
-- ============================================================================

-- ============================================================================
-- 1. SCHEDULE (from 13_appointments_scheduling.sql)
-- ============================================================================

CREATE TABLE schedule (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    active              BOOLEAN DEFAULT TRUE,
    practitioner_id     UUID NOT NULL REFERENCES practitioner(id),
    location_id         UUID REFERENCES location(id),
    service_type_code   VARCHAR(30),
    service_type_display VARCHAR(255),
    specialty_code      VARCHAR(30),
    specialty_display   VARCHAR(255),
    planning_horizon_start TIMESTAMPTZ,
    planning_horizon_end   TIMESTAMPTZ,
    comment             TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 2. SLOT (from 13_appointments_scheduling.sql)
-- ============================================================================

CREATE TABLE slot (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    schedule_id         UUID NOT NULL REFERENCES schedule(id) ON DELETE CASCADE,
    status              VARCHAR(30) NOT NULL DEFAULT 'free',
    start_time          TIMESTAMPTZ NOT NULL,
    end_time            TIMESTAMPTZ NOT NULL,
    overbooked          BOOLEAN DEFAULT FALSE,
    comment             TEXT,
    service_type_code   VARCHAR(30),
    service_type_display VARCHAR(255),
    specialty_code      VARCHAR(30),
    specialty_display   VARCHAR(255),
    appointment_type_code VARCHAR(30),
    appointment_type_display VARCHAR(255),
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 3. APPOINTMENT (from 13_appointments_scheduling.sql)
-- ============================================================================

CREATE TABLE appointment (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    status              VARCHAR(30) NOT NULL DEFAULT 'proposed',
    cancellation_reason VARCHAR(255),
    service_type_code   VARCHAR(30),
    service_type_display VARCHAR(255),
    specialty_code      VARCHAR(30),
    specialty_display   VARCHAR(255),
    appointment_type_code VARCHAR(30),
    appointment_type_display VARCHAR(255),
    priority            INTEGER,
    description         TEXT,
    start_time          TIMESTAMPTZ,
    end_time            TIMESTAMPTZ,
    minutes_duration    INTEGER,
    slot_id             UUID REFERENCES slot(id),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    practitioner_id     UUID REFERENCES practitioner(id),
    location_id         UUID REFERENCES location(id),
    reason_code         VARCHAR(30),
    reason_display      VARCHAR(255),
    reason_condition_id UUID REFERENCES condition(id),
    note                TEXT,
    patient_instruction TEXT,
    is_telehealth       BOOLEAN DEFAULT FALSE,
    telehealth_url      VARCHAR(500),
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 4. APPOINTMENT_PARTICIPANT (from 13_appointments_scheduling.sql)
-- ============================================================================

CREATE TABLE appointment_participant (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    appointment_id      UUID NOT NULL REFERENCES appointment(id) ON DELETE CASCADE,
    actor_type          VARCHAR(30) NOT NULL,
    actor_id            UUID NOT NULL,
    role_code           VARCHAR(30),
    role_display        VARCHAR(100),
    status              VARCHAR(30) NOT NULL DEFAULT 'needs-action',
    required            VARCHAR(20) DEFAULT 'required',
    period_start        TIMESTAMPTZ,
    period_end          TIMESTAMPTZ
);

-- ============================================================================
-- 5. APPOINTMENT_RESPONSE (from 13_appointments_scheduling.sql)
-- ============================================================================

CREATE TABLE appointment_response (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    appointment_id      UUID NOT NULL REFERENCES appointment(id) ON DELETE CASCADE,
    actor_type          VARCHAR(30) NOT NULL,
    actor_id            UUID NOT NULL,
    participant_status  VARCHAR(30) NOT NULL,
    comment             TEXT,
    start_time          TIMESTAMPTZ,
    end_time            TIMESTAMPTZ,
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 6. WAITLIST (from 13_appointments_scheduling.sql)
-- ============================================================================

CREATE TABLE waitlist (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    practitioner_id     UUID REFERENCES practitioner(id),
    department          VARCHAR(100),
    service_type_code   VARCHAR(30),
    service_type_display VARCHAR(255),
    priority            INTEGER DEFAULT 0,
    queue_number        INTEGER,
    status              VARCHAR(30) NOT NULL DEFAULT 'waiting',
    requested_date      TIMESTAMPTZ,
    check_in_time       TIMESTAMPTZ,
    called_time         TIMESTAMPTZ,
    completed_time      TIMESTAMPTZ,
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 7. INDEXES for T2 tables
-- ============================================================================

-- Schedule indexes
CREATE INDEX idx_schedule_fhir ON schedule(fhir_id);
CREATE INDEX idx_schedule_practitioner ON schedule(practitioner_id);
CREATE INDEX idx_schedule_active ON schedule(active) WHERE active = TRUE;

-- Slot indexes
CREATE INDEX idx_slot_fhir ON slot(fhir_id);
CREATE INDEX idx_slot_schedule ON slot(schedule_id);
CREATE INDEX idx_slot_status ON slot(status);
CREATE INDEX idx_slot_start ON slot(start_time);
CREATE INDEX idx_slot_available ON slot(status, start_time) WHERE status = 'free';

-- Appointment indexes
CREATE INDEX idx_appointment_fhir ON appointment(fhir_id);
CREATE INDEX idx_appointment_patient ON appointment(patient_id);
CREATE INDEX idx_appointment_practitioner ON appointment(practitioner_id) WHERE practitioner_id IS NOT NULL;
CREATE INDEX idx_appointment_status ON appointment(status);
CREATE INDEX idx_appointment_start ON appointment(start_time DESC) WHERE start_time IS NOT NULL;
CREATE INDEX idx_appointment_patient_start ON appointment(patient_id, start_time DESC);

-- Appointment participant indexes
CREATE INDEX idx_appt_participant_appointment ON appointment_participant(appointment_id);
CREATE INDEX idx_appt_participant_actor ON appointment_participant(actor_type, actor_id);

-- Appointment response indexes
CREATE INDEX idx_appt_response_appointment ON appointment_response(appointment_id);

-- Waitlist indexes
CREATE INDEX idx_waitlist_patient ON waitlist(patient_id);
CREATE INDEX idx_waitlist_practitioner ON waitlist(practitioner_id) WHERE practitioner_id IS NOT NULL;
CREATE INDEX idx_waitlist_department ON waitlist(department) WHERE department IS NOT NULL;
CREATE INDEX idx_waitlist_status ON waitlist(status);
CREATE INDEX idx_waitlist_queue ON waitlist(department, queue_number) WHERE status = 'waiting';
