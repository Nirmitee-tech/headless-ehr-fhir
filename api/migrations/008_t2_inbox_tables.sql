-- ============================================================================
-- T2 INBOX TABLES MIGRATION
-- Tables: message_pool, message_pool_member, inbox_message, cosign_request,
--         patient_list, patient_list_member, handoff_record
-- ============================================================================

-- ============================================================================
-- 1. MESSAGE_POOL (routing pools for InBasket messages)
-- ============================================================================

CREATE TABLE message_pool (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pool_name       VARCHAR(200) NOT NULL,
    pool_type       VARCHAR(50) NOT NULL,
    organization_id UUID REFERENCES organization(id),
    department_id   UUID,
    description     TEXT,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_message_pool_type ON message_pool(pool_type);
CREATE INDEX idx_message_pool_org ON message_pool(organization_id);

-- ============================================================================
-- 2. MESSAGE_POOL_MEMBER (pool membership)
-- ============================================================================

CREATE TABLE message_pool_member (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pool_id         UUID NOT NULL REFERENCES message_pool(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES "system_user"(id),
    role            VARCHAR(50),
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    joined_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_pool_member_pool ON message_pool_member(pool_id);
CREATE INDEX idx_pool_member_user ON message_pool_member(user_id);
CREATE UNIQUE INDEX idx_pool_member_unique ON message_pool_member(pool_id, user_id);

-- ============================================================================
-- 3. INBOX_MESSAGE (InBasket messages)
-- ============================================================================

CREATE TABLE inbox_message (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_type    VARCHAR(50) NOT NULL,
    priority        VARCHAR(20) NOT NULL DEFAULT 'normal',
    subject         VARCHAR(500) NOT NULL,
    body            TEXT,
    patient_id      UUID REFERENCES patient(id),
    encounter_id    UUID REFERENCES encounter(id),
    sender_id       UUID REFERENCES "system_user"(id),
    recipient_id    UUID REFERENCES "system_user"(id),
    pool_id         UUID REFERENCES message_pool(id),
    status          VARCHAR(20) NOT NULL DEFAULT 'unread',
    parent_id       UUID REFERENCES inbox_message(id),
    thread_id       UUID,
    is_urgent       BOOLEAN NOT NULL DEFAULT FALSE,
    due_date        TIMESTAMPTZ,
    read_at         TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_inbox_msg_recipient ON inbox_message(recipient_id);
CREATE INDEX idx_inbox_msg_sender ON inbox_message(sender_id);
CREATE INDEX idx_inbox_msg_patient ON inbox_message(patient_id);
CREATE INDEX idx_inbox_msg_pool ON inbox_message(pool_id);
CREATE INDEX idx_inbox_msg_status ON inbox_message(status);
CREATE INDEX idx_inbox_msg_type ON inbox_message(message_type);
CREATE INDEX idx_inbox_msg_thread ON inbox_message(thread_id);
CREATE INDEX idx_inbox_msg_created ON inbox_message(created_at DESC);

-- ============================================================================
-- 4. COSIGN_REQUEST (cosigning workflow)
-- ============================================================================

CREATE TABLE cosign_request (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_type   VARCHAR(100) NOT NULL,
    document_id     UUID,
    requester_id    UUID NOT NULL REFERENCES "system_user"(id),
    cosigner_id     UUID NOT NULL REFERENCES "system_user"(id),
    patient_id      UUID REFERENCES patient(id),
    encounter_id    UUID REFERENCES encounter(id),
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    note            TEXT,
    requested_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    responded_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cosign_cosigner ON cosign_request(cosigner_id);
CREATE INDEX idx_cosign_requester ON cosign_request(requester_id);
CREATE INDEX idx_cosign_status ON cosign_request(status);
CREATE INDEX idx_cosign_patient ON cosign_request(patient_id);

-- ============================================================================
-- 5. PATIENT_LIST (worklists)
-- ============================================================================

CREATE TABLE patient_list (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_name       VARCHAR(200) NOT NULL,
    list_type       VARCHAR(50) NOT NULL,
    owner_id        UUID NOT NULL REFERENCES "system_user"(id),
    department_id   UUID,
    description     TEXT,
    auto_criteria   JSONB,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_patient_list_owner ON patient_list(owner_id);
CREATE INDEX idx_patient_list_type ON patient_list(list_type);

-- ============================================================================
-- 6. PATIENT_LIST_MEMBER (patients on lists)
-- ============================================================================

CREATE TABLE patient_list_member (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id         UUID NOT NULL REFERENCES patient_list(id) ON DELETE CASCADE,
    patient_id      UUID NOT NULL REFERENCES patient(id),
    priority        INTEGER NOT NULL DEFAULT 0,
    flags           VARCHAR(200),
    one_liner       TEXT,
    added_by        UUID REFERENCES "system_user"(id),
    added_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    removed_at      TIMESTAMPTZ
);

CREATE INDEX idx_plm_list ON patient_list_member(list_id);
CREATE INDEX idx_plm_patient ON patient_list_member(patient_id);
CREATE UNIQUE INDEX idx_plm_unique ON patient_list_member(list_id, patient_id) WHERE removed_at IS NULL;

-- ============================================================================
-- 7. HANDOFF_RECORD (I-PASS/SBAR handoffs)
-- ============================================================================

CREATE TABLE handoff_record (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    encounter_id        UUID REFERENCES encounter(id),
    from_provider_id    UUID NOT NULL REFERENCES "system_user"(id),
    to_provider_id      UUID NOT NULL REFERENCES "system_user"(id),
    handoff_type        VARCHAR(30) NOT NULL DEFAULT 'ipass',
    illness_severity    VARCHAR(100),
    patient_summary     TEXT,
    action_list         TEXT,
    situation_awareness TEXT,
    synthesis           TEXT,
    contingency_plan    TEXT,
    status              VARCHAR(20) NOT NULL DEFAULT 'draft',
    acknowledged_at     TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_handoff_patient ON handoff_record(patient_id);
CREATE INDEX idx_handoff_from ON handoff_record(from_provider_id);
CREATE INDEX idx_handoff_to ON handoff_record(to_provider_id);
CREATE INDEX idx_handoff_created ON handoff_record(created_at DESC);
