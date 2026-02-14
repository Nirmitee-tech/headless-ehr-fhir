-- ============================================================================
-- T4 CDS TABLES MIGRATION
-- Tables: cds_rule, cds_alert, cds_alert_response, drug_interaction,
--         order_set, order_set_section, order_set_item,
--         clinical_pathway, clinical_pathway_phase,
--         patient_pathway_enrollment, formulary, formulary_item,
--         medication_reconciliation, medication_reconciliation_item
-- ============================================================================

-- ============================================================================
-- 1. CDS_RULE
-- ============================================================================

CREATE TABLE cds_rule (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_name           VARCHAR(255) NOT NULL,
    rule_type           VARCHAR(50) NOT NULL,
    description         TEXT,
    severity            VARCHAR(20),
    category            VARCHAR(50),
    trigger_event       VARCHAR(100),
    condition_expr      TEXT,
    action_type         VARCHAR(50),
    action_detail       TEXT,
    evidence_source     VARCHAR(255),
    evidence_url        TEXT,
    active              BOOLEAN DEFAULT TRUE,
    version             VARCHAR(30),
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 2. CDS_ALERT
-- ============================================================================

CREATE TABLE cds_alert (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id             UUID NOT NULL REFERENCES cds_rule(id),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    encounter_id        UUID REFERENCES encounter(id),
    practitioner_id     UUID REFERENCES practitioner(id),
    status              VARCHAR(30) NOT NULL DEFAULT 'fired',
    severity            VARCHAR(20),
    summary             TEXT NOT NULL,
    detail              TEXT,
    suggested_action    TEXT,
    source              VARCHAR(100),
    expires_at          TIMESTAMPTZ,
    fired_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at         TIMESTAMPTZ,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 3. CDS_ALERT_RESPONSE
-- ============================================================================

CREATE TABLE cds_alert_response (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alert_id            UUID NOT NULL REFERENCES cds_alert(id) ON DELETE CASCADE,
    practitioner_id     UUID NOT NULL REFERENCES practitioner(id),
    action              VARCHAR(30) NOT NULL,
    reason              TEXT,
    comment             TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 4. DRUG_INTERACTION
-- ============================================================================

CREATE TABLE drug_interaction (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    medication_a_id     UUID,
    medication_a_name   VARCHAR(255) NOT NULL,
    medication_b_id     UUID,
    medication_b_name   VARCHAR(255) NOT NULL,
    severity            VARCHAR(20) NOT NULL,
    description         TEXT,
    clinical_effect     TEXT,
    management          TEXT,
    evidence_level      VARCHAR(30),
    source              VARCHAR(255),
    active              BOOLEAN DEFAULT TRUE,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 5. ORDER_SET
-- ============================================================================

CREATE TABLE order_set (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                VARCHAR(255) NOT NULL,
    description         TEXT,
    category            VARCHAR(50),
    status              VARCHAR(30) NOT NULL DEFAULT 'draft',
    author_id           UUID REFERENCES practitioner(id),
    version             VARCHAR(30),
    approval_date       DATE,
    active              BOOLEAN DEFAULT TRUE,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 6. ORDER_SET_SECTION
-- ============================================================================

CREATE TABLE order_set_section (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_set_id        UUID NOT NULL REFERENCES order_set(id) ON DELETE CASCADE,
    name                VARCHAR(255) NOT NULL,
    description         TEXT,
    sort_order          INTEGER DEFAULT 0
);

-- ============================================================================
-- 7. ORDER_SET_ITEM
-- ============================================================================

CREATE TABLE order_set_item (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    section_id          UUID NOT NULL REFERENCES order_set_section(id) ON DELETE CASCADE,
    item_type           VARCHAR(30) NOT NULL,
    item_name           VARCHAR(255) NOT NULL,
    item_code           VARCHAR(50),
    default_dose        VARCHAR(100),
    default_frequency   VARCHAR(100),
    default_duration    VARCHAR(100),
    instructions        TEXT,
    is_required         BOOLEAN DEFAULT FALSE,
    sort_order          INTEGER DEFAULT 0
);

-- ============================================================================
-- 8. CLINICAL_PATHWAY
-- ============================================================================

CREATE TABLE clinical_pathway (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                VARCHAR(255) NOT NULL,
    description         TEXT,
    condition           VARCHAR(255),
    category            VARCHAR(50),
    version             VARCHAR(30),
    author_id           UUID REFERENCES practitioner(id),
    active              BOOLEAN DEFAULT TRUE,
    expected_duration   VARCHAR(50),
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 9. CLINICAL_PATHWAY_PHASE
-- ============================================================================

CREATE TABLE clinical_pathway_phase (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pathway_id          UUID NOT NULL REFERENCES clinical_pathway(id) ON DELETE CASCADE,
    name                VARCHAR(255) NOT NULL,
    description         TEXT,
    duration            VARCHAR(50),
    goals               TEXT,
    interventions       TEXT,
    sort_order          INTEGER DEFAULT 0
);

-- ============================================================================
-- 10. PATIENT_PATHWAY_ENROLLMENT
-- ============================================================================

CREATE TABLE patient_pathway_enrollment (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pathway_id          UUID NOT NULL REFERENCES clinical_pathway(id),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    practitioner_id     UUID REFERENCES practitioner(id),
    status              VARCHAR(30) NOT NULL DEFAULT 'active',
    current_phase_id    UUID REFERENCES clinical_pathway_phase(id),
    enrolled_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at        TIMESTAMPTZ,
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 11. FORMULARY
-- ============================================================================

CREATE TABLE formulary (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                VARCHAR(255) NOT NULL,
    description         TEXT,
    organization_id     UUID REFERENCES organization(id),
    effective_date      DATE,
    expiration_date     DATE,
    version             VARCHAR(30),
    active              BOOLEAN DEFAULT TRUE,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 12. FORMULARY_ITEM
-- ============================================================================

CREATE TABLE formulary_item (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    formulary_id        UUID NOT NULL REFERENCES formulary(id) ON DELETE CASCADE,
    medication_id       UUID,
    medication_name     VARCHAR(255) NOT NULL,
    tier_level          INTEGER,
    requires_prior_auth BOOLEAN DEFAULT FALSE,
    step_therapy_req    BOOLEAN DEFAULT FALSE,
    quantity_limit      VARCHAR(100),
    preferred_status    VARCHAR(30),
    note                TEXT
);

-- ============================================================================
-- 13. MEDICATION_RECONCILIATION
-- ============================================================================

CREATE TABLE medication_reconciliation (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    encounter_id        UUID REFERENCES encounter(id),
    practitioner_id     UUID REFERENCES practitioner(id),
    status              VARCHAR(30) NOT NULL DEFAULT 'in-progress',
    reconc_type         VARCHAR(30),
    completed_at        TIMESTAMPTZ,
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 14. MEDICATION_RECONCILIATION_ITEM
-- ============================================================================

CREATE TABLE medication_reconciliation_item (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reconciliation_id   UUID NOT NULL REFERENCES medication_reconciliation(id) ON DELETE CASCADE,
    medication_id       UUID,
    medication_name     VARCHAR(255) NOT NULL,
    source_list         VARCHAR(50),
    dose                VARCHAR(100),
    frequency           VARCHAR(100),
    route               VARCHAR(50),
    action              VARCHAR(30),
    reason              TEXT,
    verified_by_id      UUID REFERENCES practitioner(id),
    verified_at         TIMESTAMPTZ
);

-- ============================================================================
-- 15. INDEXES
-- ============================================================================

-- CDS Rule indexes
CREATE INDEX idx_cds_rule_name ON cds_rule(rule_name);
CREATE INDEX idx_cds_rule_type ON cds_rule(rule_type);
CREATE INDEX idx_cds_rule_active ON cds_rule(active) WHERE active = TRUE;

-- CDS Alert indexes
CREATE INDEX idx_cds_alert_rule ON cds_alert(rule_id);
CREATE INDEX idx_cds_alert_patient ON cds_alert(patient_id);
CREATE INDEX idx_cds_alert_encounter ON cds_alert(encounter_id) WHERE encounter_id IS NOT NULL;
CREATE INDEX idx_cds_alert_status ON cds_alert(status);
CREATE INDEX idx_cds_alert_fired ON cds_alert(fired_at DESC);
CREATE INDEX idx_cds_alert_patient_status ON cds_alert(patient_id, status);

-- CDS Alert Response indexes
CREATE INDEX idx_cds_alert_resp_alert ON cds_alert_response(alert_id);
CREATE INDEX idx_cds_alert_resp_practitioner ON cds_alert_response(practitioner_id);

-- Drug Interaction indexes
CREATE INDEX idx_drug_int_med_a ON drug_interaction(medication_a_id) WHERE medication_a_id IS NOT NULL;
CREATE INDEX idx_drug_int_med_b ON drug_interaction(medication_b_id) WHERE medication_b_id IS NOT NULL;
CREATE INDEX idx_drug_int_severity ON drug_interaction(severity);
CREATE INDEX idx_drug_int_active ON drug_interaction(active) WHERE active = TRUE;

-- Order Set indexes
CREATE INDEX idx_order_set_name ON order_set(name);
CREATE INDEX idx_order_set_category ON order_set(category) WHERE category IS NOT NULL;
CREATE INDEX idx_order_set_active ON order_set(active) WHERE active = TRUE;

-- Order Set Section indexes
CREATE INDEX idx_os_section_set ON order_set_section(order_set_id);
CREATE INDEX idx_os_section_sort ON order_set_section(order_set_id, sort_order);

-- Order Set Item indexes
CREATE INDEX idx_os_item_section ON order_set_item(section_id);
CREATE INDEX idx_os_item_sort ON order_set_item(section_id, sort_order);

-- Clinical Pathway indexes
CREATE INDEX idx_pathway_name ON clinical_pathway(name);
CREATE INDEX idx_pathway_condition ON clinical_pathway(condition) WHERE condition IS NOT NULL;
CREATE INDEX idx_pathway_active ON clinical_pathway(active) WHERE active = TRUE;

-- Clinical Pathway Phase indexes
CREATE INDEX idx_pathway_phase_pathway ON clinical_pathway_phase(pathway_id);
CREATE INDEX idx_pathway_phase_sort ON clinical_pathway_phase(pathway_id, sort_order);

-- Patient Pathway Enrollment indexes
CREATE INDEX idx_enrollment_pathway ON patient_pathway_enrollment(pathway_id);
CREATE INDEX idx_enrollment_patient ON patient_pathway_enrollment(patient_id);
CREATE INDEX idx_enrollment_status ON patient_pathway_enrollment(status);
CREATE INDEX idx_enrollment_patient_status ON patient_pathway_enrollment(patient_id, status);

-- Formulary indexes
CREATE INDEX idx_formulary_name ON formulary(name);
CREATE INDEX idx_formulary_org ON formulary(organization_id) WHERE organization_id IS NOT NULL;
CREATE INDEX idx_formulary_active ON formulary(active) WHERE active = TRUE;

-- Formulary Item indexes
CREATE INDEX idx_form_item_formulary ON formulary_item(formulary_id);
CREATE INDEX idx_form_item_med ON formulary_item(medication_id) WHERE medication_id IS NOT NULL;

-- Medication Reconciliation indexes
CREATE INDEX idx_med_reconc_patient ON medication_reconciliation(patient_id);
CREATE INDEX idx_med_reconc_encounter ON medication_reconciliation(encounter_id) WHERE encounter_id IS NOT NULL;
CREATE INDEX idx_med_reconc_status ON medication_reconciliation(status);
CREATE INDEX idx_med_reconc_created ON medication_reconciliation(created_at DESC);

-- Medication Reconciliation Item indexes
CREATE INDEX idx_med_reconc_item_reconc ON medication_reconciliation_item(reconciliation_id);
CREATE INDEX idx_med_reconc_item_med ON medication_reconciliation_item(medication_id) WHERE medication_id IS NOT NULL;
