-- ============================================================================
-- T4 PORTAL TABLES MIGRATION
-- Tables: portal_account, portal_proxy_access, portal_message, questionnaire,
--         questionnaire_item, questionnaire_response,
--         questionnaire_response_item, patient_checkin
-- ============================================================================

-- ============================================================================
-- 1. PORTAL_ACCOUNT
-- ============================================================================

CREATE TABLE portal_account (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    username            VARCHAR(100) UNIQUE NOT NULL,
    email               VARCHAR(255) NOT NULL,
    phone               VARCHAR(30),
    status              VARCHAR(30) NOT NULL DEFAULT 'pending-activation',
    email_verified      BOOLEAN DEFAULT FALSE,
    last_login_at       TIMESTAMPTZ,
    failed_login_count  INTEGER DEFAULT 0,
    password_last_changed TIMESTAMPTZ,
    mfa_enabled         BOOLEAN DEFAULT FALSE,
    preferred_language  VARCHAR(10),
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 2. PORTAL_PROXY_ACCESS
-- ============================================================================

CREATE TABLE portal_proxy_access (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    portal_account_id   UUID NOT NULL REFERENCES portal_account(id) ON DELETE CASCADE,
    proxy_patient_id    UUID NOT NULL REFERENCES patient(id),
    relationship        VARCHAR(30) NOT NULL,
    access_level        VARCHAR(30) NOT NULL DEFAULT 'read-only',
    active              BOOLEAN DEFAULT TRUE,
    period_start        TIMESTAMPTZ,
    period_end          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 3. PORTAL_MESSAGE
-- ============================================================================

CREATE TABLE portal_message (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    practitioner_id     UUID REFERENCES practitioner(id),
    direction           VARCHAR(10) NOT NULL,
    subject             VARCHAR(255),
    body                TEXT NOT NULL,
    status              VARCHAR(30) NOT NULL DEFAULT 'sent',
    priority            VARCHAR(20),
    category            VARCHAR(50),
    parent_id           UUID REFERENCES portal_message(id),
    read_at             TIMESTAMPTZ,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 4. QUESTIONNAIRE (FHIR Questionnaire)
-- ============================================================================

CREATE TABLE questionnaire (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    name                VARCHAR(255) NOT NULL,
    title               VARCHAR(500),
    status              VARCHAR(30) NOT NULL DEFAULT 'draft',
    version             VARCHAR(30),
    description         TEXT,
    purpose             TEXT,
    subject_type        VARCHAR(50),
    date                DATE,
    publisher           VARCHAR(255),
    approval_date       DATE,
    last_review_date    DATE,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 5. QUESTIONNAIRE_ITEM
-- ============================================================================

CREATE TABLE questionnaire_item (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    questionnaire_id    UUID NOT NULL REFERENCES questionnaire(id) ON DELETE CASCADE,
    link_id             VARCHAR(100) NOT NULL,
    text                TEXT NOT NULL,
    type                VARCHAR(30) NOT NULL,
    required            BOOLEAN DEFAULT FALSE,
    repeats             BOOLEAN DEFAULT FALSE,
    read_only           BOOLEAN DEFAULT FALSE,
    max_length          INTEGER,
    answer_options      JSONB,
    initial_value       TEXT,
    enable_when_link_id VARCHAR(100),
    enable_when_operator VARCHAR(20),
    enable_when_answer  TEXT,
    sort_order          INTEGER DEFAULT 0
);

-- ============================================================================
-- 6. QUESTIONNAIRE_RESPONSE (FHIR QuestionnaireResponse)
-- ============================================================================

CREATE TABLE questionnaire_response (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    questionnaire_id    UUID NOT NULL REFERENCES questionnaire(id),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    encounter_id        UUID REFERENCES encounter(id),
    author_id           UUID REFERENCES practitioner(id),
    status              VARCHAR(30) NOT NULL DEFAULT 'in-progress',
    authored            TIMESTAMPTZ,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 7. QUESTIONNAIRE_RESPONSE_ITEM
-- ============================================================================

CREATE TABLE questionnaire_response_item (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    response_id         UUID NOT NULL REFERENCES questionnaire_response(id) ON DELETE CASCADE,
    link_id             VARCHAR(100) NOT NULL,
    text                TEXT,
    answer_string       TEXT,
    answer_integer      INTEGER,
    answer_boolean      BOOLEAN,
    answer_date         DATE,
    answer_code         VARCHAR(100)
);

-- ============================================================================
-- 8. PATIENT_CHECKIN
-- ============================================================================

CREATE TABLE patient_checkin (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    appointment_id      UUID REFERENCES appointment(id),
    status              VARCHAR(30) NOT NULL DEFAULT 'pending',
    checkin_method      VARCHAR(30),
    checkin_time        TIMESTAMPTZ,
    insurance_verified  BOOLEAN,
    co_pay_collected    BOOLEAN,
    co_pay_amount       DECIMAL(10,2),
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 9. INDEXES
-- ============================================================================

-- Portal Account indexes
CREATE INDEX idx_portal_account_patient ON portal_account(patient_id);
CREATE INDEX idx_portal_account_username ON portal_account(username);
CREATE INDEX idx_portal_account_email ON portal_account(email);
CREATE INDEX idx_portal_account_status ON portal_account(status);

-- Portal Proxy Access indexes
CREATE INDEX idx_portal_proxy_account ON portal_proxy_access(portal_account_id);
CREATE INDEX idx_portal_proxy_patient ON portal_proxy_access(proxy_patient_id);

-- Portal Message indexes
CREATE INDEX idx_portal_msg_patient ON portal_message(patient_id);
CREATE INDEX idx_portal_msg_practitioner ON portal_message(practitioner_id) WHERE practitioner_id IS NOT NULL;
CREATE INDEX idx_portal_msg_status ON portal_message(status);
CREATE INDEX idx_portal_msg_parent ON portal_message(parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX idx_portal_msg_created ON portal_message(created_at DESC);

-- Questionnaire indexes
CREATE INDEX idx_questionnaire_fhir ON questionnaire(fhir_id);
CREATE INDEX idx_questionnaire_name ON questionnaire(name);
CREATE INDEX idx_questionnaire_status ON questionnaire(status);

-- Questionnaire Item indexes
CREATE INDEX idx_quest_item_questionnaire ON questionnaire_item(questionnaire_id);
CREATE INDEX idx_quest_item_sort ON questionnaire_item(questionnaire_id, sort_order);

-- Questionnaire Response indexes
CREATE INDEX idx_quest_resp_fhir ON questionnaire_response(fhir_id);
CREATE INDEX idx_quest_resp_questionnaire ON questionnaire_response(questionnaire_id);
CREATE INDEX idx_quest_resp_patient ON questionnaire_response(patient_id);
CREATE INDEX idx_quest_resp_status ON questionnaire_response(status);

-- Questionnaire Response Item indexes
CREATE INDEX idx_quest_resp_item_response ON questionnaire_response_item(response_id);

-- Patient Checkin indexes
CREATE INDEX idx_checkin_patient ON patient_checkin(patient_id);
CREATE INDEX idx_checkin_appointment ON patient_checkin(appointment_id) WHERE appointment_id IS NOT NULL;
CREATE INDEX idx_checkin_status ON patient_checkin(status);
CREATE INDEX idx_checkin_time ON patient_checkin(checkin_time DESC) WHERE checkin_time IS NOT NULL;
