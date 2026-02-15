-- 023_fhir_resource_gaps.sql
-- Adds all missing FHIR R4 resource tables to close coverage gaps.
-- Includes clinical safety, care delivery, financial, workflow,
-- supply, conformance/messaging, and specialty resource types.

-- ============================================================
-- Clinical Safety Tables
-- ============================================================

-- Flag (patient alerts/warnings)
CREATE TABLE IF NOT EXISTS flag (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'active',
    category_code   TEXT,
    category_display TEXT,
    code_code       TEXT NOT NULL,
    code_display    TEXT,
    code_system     TEXT,
    subject_patient_id UUID REFERENCES patient(id),
    period_start    TIMESTAMPTZ,
    period_end      TIMESTAMPTZ,
    encounter_id    UUID REFERENCES encounter(id),
    author_practitioner_id UUID REFERENCES practitioner(id),
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_flag_patient ON flag (subject_patient_id);
CREATE INDEX IF NOT EXISTS idx_flag_status ON flag (status);

-- fhir_list (FHIR List resource, named fhir_list to avoid SQL keyword)
CREATE TABLE IF NOT EXISTS fhir_list (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'current',
    mode            TEXT NOT NULL DEFAULT 'working',
    title           TEXT,
    code_code       TEXT,
    code_display    TEXT,
    subject_patient_id UUID REFERENCES patient(id),
    encounter_id    UUID REFERENCES encounter(id),
    date            TIMESTAMPTZ,
    source_practitioner_id UUID REFERENCES practitioner(id),
    ordered_by      TEXT,
    note            TEXT,
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_fhir_list_patient ON fhir_list (subject_patient_id);
CREATE INDEX IF NOT EXISTS idx_fhir_list_status ON fhir_list (status);

CREATE TABLE IF NOT EXISTS fhir_list_entry (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id         UUID NOT NULL REFERENCES fhir_list(id) ON DELETE CASCADE,
    item_reference  TEXT NOT NULL,
    item_display    TEXT,
    date            TIMESTAMPTZ,
    deleted         BOOLEAN DEFAULT false,
    flag_code       TEXT,
    flag_display    TEXT
);
CREATE INDEX IF NOT EXISTS idx_fhir_list_entry_list ON fhir_list_entry (list_id);

-- DetectedIssue (clinical decision support alerts)
CREATE TABLE IF NOT EXISTS detected_issue (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'final',
    code_code       TEXT,
    code_display    TEXT,
    code_system     TEXT,
    severity        TEXT,
    patient_id      UUID REFERENCES patient(id),
    identified_date TIMESTAMPTZ,
    author_practitioner_id UUID REFERENCES practitioner(id),
    implicated      JSONB,
    detail          TEXT,
    reference_url   TEXT,
    mitigation_action TEXT,
    mitigation_date TIMESTAMPTZ,
    mitigation_author_id UUID REFERENCES practitioner(id),
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_detected_issue_patient ON detected_issue (patient_id);
CREATE INDEX IF NOT EXISTS idx_detected_issue_status ON detected_issue (status);

-- AdverseEvent (patient safety reporting)
CREATE TABLE IF NOT EXISTS adverse_event (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT NOT NULL UNIQUE,
    actuality       TEXT NOT NULL DEFAULT 'actual',
    category_code   TEXT,
    category_display TEXT,
    event_code      TEXT,
    event_display   TEXT,
    event_system    TEXT,
    subject_patient_id UUID REFERENCES patient(id) NOT NULL,
    encounter_id    UUID REFERENCES encounter(id),
    date            TIMESTAMPTZ,
    detected        TIMESTAMPTZ,
    recorded_date   TIMESTAMPTZ,
    recorder_id     UUID REFERENCES practitioner(id),
    seriousness_code TEXT,
    severity_code   TEXT,
    outcome_code    TEXT,
    location_id     UUID REFERENCES location(id),
    description     TEXT,
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_adverse_event_patient ON adverse_event (subject_patient_id);
CREATE INDEX IF NOT EXISTS idx_adverse_event_actuality ON adverse_event (actuality);

-- ClinicalImpression (clinical reasoning documentation)
CREATE TABLE IF NOT EXISTS clinical_impression (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'completed',
    status_reason   TEXT,
    code_code       TEXT,
    code_display    TEXT,
    description     TEXT,
    subject_patient_id UUID REFERENCES patient(id) NOT NULL,
    encounter_id    UUID REFERENCES encounter(id),
    effective_date  TIMESTAMPTZ,
    date            TIMESTAMPTZ,
    assessor_id     UUID REFERENCES practitioner(id),
    previous_id     UUID,
    problem         JSONB,
    summary         TEXT,
    finding         JSONB,
    prognosis_code  TEXT,
    prognosis_display TEXT,
    note            TEXT,
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_clinical_impression_patient ON clinical_impression (subject_patient_id);
CREATE INDEX IF NOT EXISTS idx_clinical_impression_status ON clinical_impression (status);

-- RiskAssessment (structured risk scores)
CREATE TABLE IF NOT EXISTS risk_assessment (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'final',
    method_code     TEXT,
    method_display  TEXT,
    code_code       TEXT,
    code_display    TEXT,
    subject_patient_id UUID REFERENCES patient(id) NOT NULL,
    encounter_id    UUID REFERENCES encounter(id),
    occurrence_date TIMESTAMPTZ,
    condition_id    UUID,
    performer_id    UUID REFERENCES practitioner(id),
    basis           JSONB,
    prediction_outcome TEXT,
    prediction_probability NUMERIC,
    prediction_qualitative TEXT,
    prediction_when_start TIMESTAMPTZ,
    prediction_when_end TIMESTAMPTZ,
    mitigation      TEXT,
    note            TEXT,
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_risk_assessment_patient ON risk_assessment (subject_patient_id);
CREATE INDEX IF NOT EXISTS idx_risk_assessment_status ON risk_assessment (status);

-- ============================================================
-- Care Delivery Tables
-- ============================================================

-- EpisodeOfCare
CREATE TABLE IF NOT EXISTS episode_of_care (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'active',
    type_code       TEXT,
    type_display    TEXT,
    diagnosis_condition_id UUID,
    diagnosis_role  TEXT,
    patient_id      UUID REFERENCES patient(id) NOT NULL,
    managing_org_id UUID REFERENCES organization(id),
    period_start    TIMESTAMPTZ,
    period_end      TIMESTAMPTZ,
    referral_request_id UUID,
    care_manager_id UUID REFERENCES practitioner(id),
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_episode_of_care_patient ON episode_of_care (patient_id);
CREATE INDEX IF NOT EXISTS idx_episode_of_care_status ON episode_of_care (status);

-- HealthcareService
CREATE TABLE IF NOT EXISTS healthcare_service (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT NOT NULL UNIQUE,
    active          BOOLEAN DEFAULT true,
    provided_by_org_id UUID REFERENCES organization(id),
    category_code   TEXT,
    category_display TEXT,
    type_code       TEXT,
    type_display    TEXT,
    name            TEXT,
    comment         TEXT,
    telecom_phone   TEXT,
    telecom_email   TEXT,
    service_provision_code TEXT,
    program_name    TEXT,
    location_id     UUID REFERENCES location(id),
    appointment_required BOOLEAN,
    available_time  JSONB,
    not_available   JSONB,
    availability_exceptions TEXT,
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_healthcare_service_org ON healthcare_service (provided_by_org_id);
CREATE INDEX IF NOT EXISTS idx_healthcare_service_active ON healthcare_service (active);

-- MeasureReport — extend existing table from migration 020 with FHIR R4 columns
-- Migration 020 already created measure_report with a minimal schema; add FHIR columns if missing.
DO $$ BEGIN
    ALTER TABLE measure_report ADD COLUMN IF NOT EXISTS fhir_id TEXT;
    ALTER TABLE measure_report ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'complete';
    ALTER TABLE measure_report ADD COLUMN IF NOT EXISTS type TEXT NOT NULL DEFAULT 'individual';
    ALTER TABLE measure_report ADD COLUMN IF NOT EXISTS measure_url TEXT;
    ALTER TABLE measure_report ADD COLUMN IF NOT EXISTS subject_patient_id UUID REFERENCES patient(id);
    ALTER TABLE measure_report ADD COLUMN IF NOT EXISTS date TIMESTAMPTZ;
    ALTER TABLE measure_report ADD COLUMN IF NOT EXISTS reporter_org_id UUID REFERENCES organization(id);
    ALTER TABLE measure_report ADD COLUMN IF NOT EXISTS improvement_notation TEXT;
    ALTER TABLE measure_report ADD COLUMN IF NOT EXISTS group_code TEXT;
    ALTER TABLE measure_report ADD COLUMN IF NOT EXISTS group_population_code TEXT;
    ALTER TABLE measure_report ADD COLUMN IF NOT EXISTS group_population_count INTEGER;
    ALTER TABLE measure_report ADD COLUMN IF NOT EXISTS group_measure_score NUMERIC;
    ALTER TABLE measure_report ADD COLUMN IF NOT EXISTS version_id INTEGER NOT NULL DEFAULT 1;
    ALTER TABLE measure_report ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;
-- Add unique constraint on fhir_id if not already present
DO $$ BEGIN
    ALTER TABLE measure_report ADD CONSTRAINT measure_report_fhir_id_unique UNIQUE (fhir_id);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
CREATE INDEX IF NOT EXISTS idx_measure_report_patient ON measure_report (subject_patient_id);
CREATE INDEX IF NOT EXISTS idx_measure_report_status ON measure_report (status);

-- ============================================================
-- Financial Tables
-- ============================================================

-- Account (billing account)
CREATE TABLE IF NOT EXISTS account (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'active',
    type_code       TEXT,
    type_display    TEXT,
    name            TEXT,
    subject_patient_id UUID REFERENCES patient(id),
    service_period_start TIMESTAMPTZ,
    service_period_end TIMESTAMPTZ,
    owner_org_id    UUID REFERENCES organization(id),
    description     TEXT,
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_account_patient ON account (subject_patient_id);
CREATE INDEX IF NOT EXISTS idx_account_status ON account (status);

-- InsurancePlan
CREATE TABLE IF NOT EXISTS insurance_plan (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'active',
    type_code       TEXT,
    type_display    TEXT,
    name            TEXT,
    alias           TEXT,
    period_start    TIMESTAMPTZ,
    period_end      TIMESTAMPTZ,
    owned_by_org_id UUID REFERENCES organization(id),
    administered_by_org_id UUID REFERENCES organization(id),
    coverage_area   TEXT,
    network_name    TEXT,
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_insurance_plan_status ON insurance_plan (status);

-- PaymentNotice
CREATE TABLE IF NOT EXISTS payment_notice (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'active',
    request_reference TEXT,
    response_reference TEXT,
    created         TIMESTAMPTZ,
    provider_id     UUID REFERENCES practitioner(id),
    payment_reference TEXT,
    payment_date    DATE,
    payee_org_id    UUID REFERENCES organization(id),
    recipient_org_id UUID REFERENCES organization(id),
    amount_value    NUMERIC,
    amount_currency TEXT DEFAULT 'USD',
    payment_status_code TEXT,
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_payment_notice_status ON payment_notice (status);

-- PaymentReconciliation
CREATE TABLE IF NOT EXISTS payment_reconciliation (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'active',
    period_start    TIMESTAMPTZ,
    period_end      TIMESTAMPTZ,
    created         TIMESTAMPTZ,
    payment_issuer_org_id UUID REFERENCES organization(id),
    request_reference TEXT,
    requestor_id    UUID REFERENCES practitioner(id),
    outcome         TEXT,
    disposition     TEXT,
    payment_date    DATE NOT NULL,
    payment_amount  NUMERIC NOT NULL,
    payment_currency TEXT DEFAULT 'USD',
    payment_identifier TEXT,
    form_code       TEXT,
    process_note    TEXT,
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_payment_reconciliation_status ON payment_reconciliation (status);

-- ChargeItem
CREATE TABLE IF NOT EXISTS charge_item (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'billable',
    code_code       TEXT NOT NULL,
    code_display    TEXT,
    code_system     TEXT,
    subject_patient_id UUID REFERENCES patient(id) NOT NULL,
    context_encounter_id UUID REFERENCES encounter(id),
    occurrence_date TIMESTAMPTZ,
    performer_id    UUID REFERENCES practitioner(id),
    performing_org_id UUID REFERENCES organization(id),
    quantity_value  NUMERIC,
    factor_override NUMERIC,
    price_override_value NUMERIC,
    price_override_currency TEXT DEFAULT 'USD',
    override_reason TEXT,
    enterer_id      UUID REFERENCES practitioner(id),
    entered_date    TIMESTAMPTZ,
    account_id      UUID REFERENCES account(id),
    note            TEXT,
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_charge_item_patient ON charge_item (subject_patient_id);
CREATE INDEX IF NOT EXISTS idx_charge_item_status ON charge_item (status);

-- ChargeItemDefinition
CREATE TABLE IF NOT EXISTS charge_item_definition (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT NOT NULL UNIQUE,
    url             TEXT,
    status          TEXT NOT NULL DEFAULT 'active',
    title           TEXT,
    description     TEXT,
    code_code       TEXT,
    code_display    TEXT,
    code_system     TEXT,
    effective_start TIMESTAMPTZ,
    effective_end   TIMESTAMPTZ,
    publisher       TEXT,
    approval_date   DATE,
    last_review_date DATE,
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_charge_item_def_status ON charge_item_definition (status);

-- Contract
CREATE TABLE IF NOT EXISTS contract (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'executed',
    type_code       TEXT,
    type_display    TEXT,
    sub_type_code   TEXT,
    title           TEXT,
    issued          TIMESTAMPTZ,
    applies_start   TIMESTAMPTZ,
    applies_end     TIMESTAMPTZ,
    subject_patient_id UUID REFERENCES patient(id),
    authority_org_id UUID REFERENCES organization(id),
    scope_code      TEXT,
    scope_display   TEXT,
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_contract_status ON contract (status);
CREATE INDEX IF NOT EXISTS idx_contract_patient ON contract (subject_patient_id);

-- EnrollmentRequest
CREATE TABLE IF NOT EXISTS enrollment_request (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'active',
    created         TIMESTAMPTZ,
    insurer_org_id  UUID REFERENCES organization(id),
    provider_id     UUID REFERENCES practitioner(id),
    candidate_patient_id UUID REFERENCES patient(id),
    coverage_id     UUID REFERENCES coverage(id),
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_enrollment_request_status ON enrollment_request (status);

-- EnrollmentResponse
CREATE TABLE IF NOT EXISTS enrollment_response (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'active',
    request_id      UUID REFERENCES enrollment_request(id),
    outcome         TEXT,
    disposition     TEXT,
    created         TIMESTAMPTZ,
    organization_id UUID REFERENCES organization(id),
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_enrollment_response_status ON enrollment_response (status);

-- ============================================================
-- Workflow Tables
-- ============================================================

-- ActivityDefinition
CREATE TABLE IF NOT EXISTS activity_definition (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT NOT NULL UNIQUE,
    url             TEXT,
    status          TEXT NOT NULL DEFAULT 'active',
    name            TEXT,
    title           TEXT,
    description     TEXT,
    purpose         TEXT,
    kind            TEXT,
    code_code       TEXT,
    code_display    TEXT,
    code_system     TEXT,
    intent          TEXT,
    priority        TEXT,
    do_not_perform  BOOLEAN DEFAULT false,
    timing_description TEXT,
    location_id     UUID REFERENCES location(id),
    quantity_value  NUMERIC,
    quantity_unit   TEXT,
    dosage_text     TEXT,
    publisher       TEXT,
    effective_start TIMESTAMPTZ,
    effective_end   TIMESTAMPTZ,
    approval_date   DATE,
    last_review_date DATE,
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_activity_definition_status ON activity_definition (status);

-- RequestGroup
CREATE TABLE IF NOT EXISTS request_group (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'active',
    intent          TEXT NOT NULL DEFAULT 'proposal',
    priority        TEXT,
    code_code       TEXT,
    code_display    TEXT,
    subject_patient_id UUID REFERENCES patient(id),
    encounter_id    UUID REFERENCES encounter(id),
    authored_on     TIMESTAMPTZ,
    author_id       UUID REFERENCES practitioner(id),
    reason_code     TEXT,
    reason_display  TEXT,
    note            TEXT,
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_request_group_patient ON request_group (subject_patient_id);
CREATE INDEX IF NOT EXISTS idx_request_group_status ON request_group (status);

CREATE TABLE IF NOT EXISTS request_group_action (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_group_id UUID NOT NULL REFERENCES request_group(id) ON DELETE CASCADE,
    prefix          TEXT,
    title           TEXT,
    description     TEXT,
    priority        TEXT,
    resource_reference TEXT,
    selection_behavior TEXT,
    required_behavior TEXT,
    precheck_behavior TEXT,
    cardinality_behavior TEXT
);
CREATE INDEX IF NOT EXISTS idx_request_group_action_group ON request_group_action (request_group_id);

-- GuidanceResponse
CREATE TABLE IF NOT EXISTS guidance_response (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT NOT NULL UNIQUE,
    request_identifier TEXT,
    module_uri      TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'success',
    subject_patient_id UUID REFERENCES patient(id),
    encounter_id    UUID REFERENCES encounter(id),
    occurrence_date TIMESTAMPTZ,
    performer_id    UUID REFERENCES practitioner(id),
    reason_code     TEXT,
    reason_display  TEXT,
    note            TEXT,
    data_requirement JSONB,
    result_reference TEXT,
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_guidance_response_patient ON guidance_response (subject_patient_id);
CREATE INDEX IF NOT EXISTS idx_guidance_response_status ON guidance_response (status);

-- ============================================================
-- Supply Tables
-- ============================================================

-- SupplyRequest
CREATE TABLE IF NOT EXISTS supply_request (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'active',
    category_code   TEXT,
    category_display TEXT,
    priority        TEXT,
    item_code       TEXT NOT NULL,
    item_display    TEXT,
    item_system     TEXT,
    quantity_value  NUMERIC NOT NULL,
    quantity_unit   TEXT,
    occurrence_date TIMESTAMPTZ,
    authored_on     TIMESTAMPTZ,
    requester_id    UUID REFERENCES practitioner(id),
    supplier_org_id UUID REFERENCES organization(id),
    deliver_to_location_id UUID REFERENCES location(id),
    reason_code     TEXT,
    reason_display  TEXT,
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_supply_request_status ON supply_request (status);

-- SupplyDelivery
CREATE TABLE IF NOT EXISTS supply_delivery (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'completed',
    based_on_id     UUID REFERENCES supply_request(id),
    patient_id      UUID REFERENCES patient(id),
    type_code       TEXT,
    type_display    TEXT,
    supplied_item_code TEXT,
    supplied_item_display TEXT,
    supplied_item_quantity NUMERIC,
    supplied_item_unit TEXT,
    occurrence_date TIMESTAMPTZ,
    supplier_id     UUID REFERENCES practitioner(id),
    destination_location_id UUID REFERENCES location(id),
    receiver_id     UUID REFERENCES practitioner(id),
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_supply_delivery_status ON supply_delivery (status);
CREATE INDEX IF NOT EXISTS idx_supply_delivery_based_on ON supply_delivery (based_on_id);

-- ============================================================
-- Conformance / Messaging Tables
-- ============================================================

-- NamingSystem
CREATE TABLE IF NOT EXISTS naming_system (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT NOT NULL UNIQUE,
    name            TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'active',
    kind            TEXT NOT NULL DEFAULT 'identifier',
    date            TIMESTAMPTZ,
    publisher       TEXT,
    responsible     TEXT,
    type_code       TEXT,
    type_display    TEXT,
    description     TEXT,
    usage_note      TEXT,
    jurisdiction    TEXT,
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_naming_system_status ON naming_system (status);

CREATE TABLE IF NOT EXISTS naming_system_unique_id (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    naming_system_id UUID NOT NULL REFERENCES naming_system(id) ON DELETE CASCADE,
    type            TEXT NOT NULL,
    value           TEXT NOT NULL,
    preferred       BOOLEAN DEFAULT false,
    comment         TEXT,
    period_start    TIMESTAMPTZ,
    period_end      TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_naming_system_uid_ns ON naming_system_unique_id (naming_system_id);

-- OperationDefinition
CREATE TABLE IF NOT EXISTS operation_definition (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT NOT NULL UNIQUE,
    url             TEXT,
    name            TEXT NOT NULL,
    title           TEXT,
    status          TEXT NOT NULL DEFAULT 'active',
    kind            TEXT NOT NULL DEFAULT 'operation',
    description     TEXT,
    code            TEXT NOT NULL,
    system          BOOLEAN DEFAULT false,
    type            BOOLEAN DEFAULT false,
    instance        BOOLEAN DEFAULT false,
    input_profile   TEXT,
    output_profile  TEXT,
    publisher       TEXT,
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_operation_definition_status ON operation_definition (status);

CREATE TABLE IF NOT EXISTS operation_definition_parameter (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    operation_definition_id UUID NOT NULL REFERENCES operation_definition(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    use             TEXT NOT NULL,
    min_val         INTEGER DEFAULT 0,
    max_val         TEXT DEFAULT '*',
    documentation   TEXT,
    type            TEXT,
    search_type     TEXT
);
CREATE INDEX IF NOT EXISTS idx_op_def_param_op ON operation_definition_parameter (operation_definition_id);

-- MessageDefinition
CREATE TABLE IF NOT EXISTS message_definition (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT NOT NULL UNIQUE,
    url             TEXT,
    name            TEXT,
    title           TEXT,
    status          TEXT NOT NULL DEFAULT 'active',
    date            TIMESTAMPTZ,
    publisher       TEXT,
    description     TEXT,
    purpose         TEXT,
    event_coding_code TEXT NOT NULL,
    event_coding_system TEXT,
    event_coding_display TEXT,
    category        TEXT,
    response_required TEXT,
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_message_definition_status ON message_definition (status);

-- MessageHeader
CREATE TABLE IF NOT EXISTS message_header (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT NOT NULL UNIQUE,
    event_coding_code TEXT NOT NULL,
    event_coding_system TEXT,
    event_coding_display TEXT,
    destination_name TEXT,
    destination_endpoint TEXT,
    sender_org_id   UUID REFERENCES organization(id),
    source_name     TEXT,
    source_endpoint TEXT NOT NULL,
    source_software TEXT,
    source_version  TEXT,
    reason_code     TEXT,
    reason_display  TEXT,
    response_identifier TEXT,
    response_code   TEXT,
    focus_reference TEXT,
    definition_url  TEXT,
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_message_header_event ON message_header (event_coding_code);

-- ============================================================
-- Specialty Tables
-- ============================================================

-- VisionPrescription
CREATE TABLE IF NOT EXISTS vision_prescription (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'active',
    created         TIMESTAMPTZ,
    patient_id      UUID REFERENCES patient(id) NOT NULL,
    encounter_id    UUID REFERENCES encounter(id),
    date_written    TIMESTAMPTZ,
    prescriber_id   UUID REFERENCES practitioner(id),
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_vision_prescription_patient ON vision_prescription (patient_id);
CREATE INDEX IF NOT EXISTS idx_vision_prescription_status ON vision_prescription (status);

CREATE TABLE IF NOT EXISTS vision_prescription_lensspec (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    prescription_id UUID NOT NULL REFERENCES vision_prescription(id) ON DELETE CASCADE,
    product_code    TEXT NOT NULL,
    product_display TEXT,
    eye             TEXT NOT NULL,
    sphere          NUMERIC,
    cylinder        NUMERIC,
    axis            INTEGER,
    prism_amount    NUMERIC,
    prism_base      TEXT,
    add_power       NUMERIC,
    power           NUMERIC,
    back_curve      NUMERIC,
    diameter        NUMERIC,
    duration_value  NUMERIC,
    duration_unit   TEXT,
    color           TEXT,
    brand           TEXT,
    note            TEXT
);
CREATE INDEX IF NOT EXISTS idx_vision_lensspec_prescription ON vision_prescription_lensspec (prescription_id);
