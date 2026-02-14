-- ============================================================================
-- T3 SURGERY TABLES MIGRATION
-- Tables: or_room, surgical_case, surgical_case_procedure, surgical_case_team,
--         surgical_time_event, surgical_preference_card, surgical_count,
--         implant_log, surgical_supply_used
-- ============================================================================

-- ============================================================================
-- 1. OR ROOM
-- ============================================================================

CREATE TABLE or_room (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                VARCHAR(100) NOT NULL,
    location_id         UUID REFERENCES location(id),
    status              VARCHAR(20) NOT NULL DEFAULT 'available',
    room_type           VARCHAR(50),
    equipment           TEXT,
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,
    decontaminated_at   TIMESTAMPTZ,
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 2. SURGICAL CASE
-- ============================================================================

CREATE TABLE surgical_case (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id            UUID NOT NULL REFERENCES patient(id),
    encounter_id          UUID REFERENCES encounter(id),
    primary_surgeon_id    UUID NOT NULL REFERENCES practitioner(id),
    anesthesiologist_id   UUID REFERENCES practitioner(id),
    or_room_id            UUID REFERENCES or_room(id),
    status                VARCHAR(20) NOT NULL DEFAULT 'scheduled',
    case_class            VARCHAR(30),
    asa_class             VARCHAR(10),
    wound_class           VARCHAR(30),
    scheduled_date        DATE NOT NULL,
    scheduled_start       TIMESTAMPTZ,
    scheduled_end         TIMESTAMPTZ,
    actual_start          TIMESTAMPTZ,
    actual_end            TIMESTAMPTZ,
    anesthesia_type       VARCHAR(30),
    laterality            VARCHAR(20),
    pre_op_diagnosis      TEXT,
    post_op_diagnosis     TEXT,
    cancel_reason         TEXT,
    note                  TEXT,
    created_at            TIMESTAMPTZ DEFAULT NOW(),
    updated_at            TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 3. SURGICAL CASE PROCEDURE
-- ============================================================================

CREATE TABLE surgical_case_procedure (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    surgical_case_id    UUID NOT NULL REFERENCES surgical_case(id) ON DELETE CASCADE,
    procedure_code      VARCHAR(30) NOT NULL,
    procedure_display   VARCHAR(500) NOT NULL,
    code_system         VARCHAR(255),
    cpt_code            VARCHAR(10),
    is_primary          BOOLEAN NOT NULL DEFAULT FALSE,
    body_site_code      VARCHAR(30),
    body_site_display   VARCHAR(255),
    sequence            INTEGER NOT NULL DEFAULT 1
);

-- ============================================================================
-- 4. SURGICAL CASE TEAM
-- ============================================================================

CREATE TABLE surgical_case_team (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    surgical_case_id    UUID NOT NULL REFERENCES surgical_case(id) ON DELETE CASCADE,
    practitioner_id     UUID NOT NULL REFERENCES practitioner(id),
    role                VARCHAR(50) NOT NULL,
    role_display        VARCHAR(100),
    start_time          TIMESTAMPTZ,
    end_time            TIMESTAMPTZ
);

-- ============================================================================
-- 5. SURGICAL TIME EVENT
-- ============================================================================

CREATE TABLE surgical_time_event (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    surgical_case_id    UUID NOT NULL REFERENCES surgical_case(id) ON DELETE CASCADE,
    event_type          VARCHAR(50) NOT NULL,
    event_time          TIMESTAMPTZ NOT NULL,
    recorded_by         UUID REFERENCES practitioner(id),
    note                TEXT
);

-- ============================================================================
-- 6. SURGICAL PREFERENCE CARD
-- ============================================================================

CREATE TABLE surgical_preference_card (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    surgeon_id          UUID NOT NULL REFERENCES practitioner(id),
    procedure_code      VARCHAR(30) NOT NULL,
    procedure_display   VARCHAR(500) NOT NULL,
    glove_size_l        VARCHAR(10),
    glove_size_r        VARCHAR(10),
    gown                VARCHAR(50),
    skin_prep           VARCHAR(100),
    position            VARCHAR(100),
    instruments         TEXT,
    supplies            TEXT,
    sutures             TEXT,
    dressings           TEXT,
    special_equipment   TEXT,
    note                TEXT,
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 7. SURGICAL COUNT
-- ============================================================================

CREATE TABLE surgical_count (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    surgical_case_id    UUID NOT NULL REFERENCES surgical_case(id) ON DELETE CASCADE,
    count_type          VARCHAR(30) NOT NULL,
    item_name           VARCHAR(255) NOT NULL,
    expected_count      INTEGER NOT NULL DEFAULT 0,
    actual_count        INTEGER NOT NULL DEFAULT 0,
    is_correct          BOOLEAN NOT NULL DEFAULT TRUE,
    counted_by          UUID REFERENCES practitioner(id),
    verified_by         UUID REFERENCES practitioner(id),
    count_time          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    note                TEXT
);

-- ============================================================================
-- 8. IMPLANT LOG
-- ============================================================================

CREATE TABLE implant_log (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    surgical_case_id    UUID REFERENCES surgical_case(id),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    device_id           UUID,
    implant_type        VARCHAR(100) NOT NULL,
    manufacturer        VARCHAR(255),
    model_number        VARCHAR(100),
    serial_number       VARCHAR(100),
    lot_number          VARCHAR(100),
    expiration_date     DATE,
    body_site_code      VARCHAR(30),
    body_site_display   VARCHAR(255),
    implanted_by        UUID REFERENCES practitioner(id),
    implant_date        TIMESTAMPTZ,
    explant_date        TIMESTAMPTZ,
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 9. SURGICAL SUPPLY USED
-- ============================================================================

CREATE TABLE surgical_supply_used (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    surgical_case_id    UUID NOT NULL REFERENCES surgical_case(id) ON DELETE CASCADE,
    supply_name         VARCHAR(255) NOT NULL,
    supply_code         VARCHAR(30),
    quantity            INTEGER NOT NULL DEFAULT 1,
    unit_of_measure     VARCHAR(30),
    lot_number          VARCHAR(100),
    recorded_by         UUID REFERENCES practitioner(id),
    note                TEXT
);

-- ============================================================================
-- 10. INDEXES for T3 Surgery tables
-- ============================================================================

-- OR Room indexes
CREATE INDEX idx_or_room_status ON or_room(status);
CREATE INDEX idx_or_room_active ON or_room(is_active) WHERE is_active = TRUE;

-- Surgical Case indexes
CREATE INDEX idx_surgical_case_patient ON surgical_case(patient_id);
CREATE INDEX idx_surgical_case_surgeon ON surgical_case(primary_surgeon_id);
CREATE INDEX idx_surgical_case_status ON surgical_case(status);
CREATE INDEX idx_surgical_case_date ON surgical_case(scheduled_date DESC);
CREATE INDEX idx_surgical_case_or_room ON surgical_case(or_room_id) WHERE or_room_id IS NOT NULL;
CREATE INDEX idx_surgical_case_encounter ON surgical_case(encounter_id) WHERE encounter_id IS NOT NULL;

-- Surgical Case Procedure indexes
CREATE INDEX idx_surg_proc_case ON surgical_case_procedure(surgical_case_id);
CREATE INDEX idx_surg_proc_code ON surgical_case_procedure(procedure_code);

-- Surgical Case Team indexes
CREATE INDEX idx_surg_team_case ON surgical_case_team(surgical_case_id);
CREATE INDEX idx_surg_team_practitioner ON surgical_case_team(practitioner_id);

-- Surgical Time Event indexes
CREATE INDEX idx_surg_time_case ON surgical_time_event(surgical_case_id);
CREATE INDEX idx_surg_time_type ON surgical_time_event(event_type);

-- Preference Card indexes
CREATE INDEX idx_pref_card_surgeon ON surgical_preference_card(surgeon_id);
CREATE INDEX idx_pref_card_procedure ON surgical_preference_card(procedure_code);

-- Surgical Count indexes
CREATE INDEX idx_surg_count_case ON surgical_count(surgical_case_id);

-- Implant Log indexes
CREATE INDEX idx_implant_patient ON implant_log(patient_id);
CREATE INDEX idx_implant_case ON implant_log(surgical_case_id) WHERE surgical_case_id IS NOT NULL;
CREATE INDEX idx_implant_serial ON implant_log(serial_number) WHERE serial_number IS NOT NULL;

-- Surgical Supply Used indexes
CREATE INDEX idx_surg_supply_case ON surgical_supply_used(surgical_case_id);
