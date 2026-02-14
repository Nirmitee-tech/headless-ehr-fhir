-- ============================================================================
-- T2 BILLING TABLES MIGRATION
-- Tables: coverage, claim, claim_diagnosis, claim_procedure, claim_item,
--         claim_response, claim_response_item, explanation_of_benefit,
--         invoice, invoice_line_item
-- ============================================================================

-- ============================================================================
-- 1. COVERAGE (Insurance Policies)
-- ============================================================================

CREATE TABLE coverage (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    status              VARCHAR(20) NOT NULL,
    type_code           VARCHAR(30),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    subscriber_id       VARCHAR(100),
    subscriber_name     VARCHAR(255),
    subscriber_dob      DATE,
    relationship        VARCHAR(30),
    dependent_number    VARCHAR(10),
    payor_org_id        UUID REFERENCES organization(id),
    payor_name          VARCHAR(255),
    policy_number       VARCHAR(100),
    group_number        VARCHAR(100),
    group_name          VARCHAR(255),
    plan_name           VARCHAR(255),
    plan_type           VARCHAR(30),
    -- US Specific
    member_id           VARCHAR(100),
    bin_number          VARCHAR(20),
    pcn_number          VARCHAR(20),
    rx_group            VARCHAR(50),
    plan_type_us        VARCHAR(20),
    -- India Specific
    ab_pmjay_id         VARCHAR(50),
    ab_pmjay_family_id  VARCHAR(50),
    state_scheme_id     VARCHAR(50),
    state_scheme_name   VARCHAR(100),
    esis_number         VARCHAR(30),
    cghs_beneficiary_id VARCHAR(30),
    echs_card_number    VARCHAR(30),
    -- Period
    period_start        DATE,
    period_end          DATE,
    -- Benefits Summary
    network             VARCHAR(20),
    copay_amount        DECIMAL(10,2),
    copay_percentage    DECIMAL(5,2),
    deductible_amount   DECIMAL(10,2),
    deductible_met      DECIMAL(10,2),
    max_benefit_amount  DECIMAL(12,2),
    out_of_pocket_max   DECIMAL(12,2),
    currency            VARCHAR(3) DEFAULT 'USD',
    coverage_order      INTEGER DEFAULT 1,
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 2. CLAIM (Billing Claims)
-- ============================================================================

CREATE TABLE claim (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    status              VARCHAR(20) NOT NULL,
    type_code           VARCHAR(30),
    sub_type_code       VARCHAR(30),
    use_code            VARCHAR(20),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    encounter_id        UUID REFERENCES encounter(id),
    insurer_org_id      UUID REFERENCES organization(id),
    provider_id         UUID REFERENCES practitioner(id),
    provider_org_id     UUID REFERENCES organization(id),
    coverage_id         UUID REFERENCES coverage(id),
    priority_code       VARCHAR(20),
    prescription_id     UUID REFERENCES medication_request(id),
    referral_id         UUID,
    facility_id         UUID REFERENCES location(id),
    billable_period_start DATE,
    billable_period_end   DATE,
    created_date        TIMESTAMPTZ DEFAULT NOW(),
    total_amount        DECIMAL(12,2),
    currency            VARCHAR(3) DEFAULT 'USD',
    place_of_service    VARCHAR(5),
    ab_pmjay_claim_id   VARCHAR(50),
    ab_pmjay_package_code VARCHAR(30),
    rohini_claim_id     VARCHAR(50),
    related_claim_id    UUID REFERENCES claim(id),
    related_claim_relationship VARCHAR(20),
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 3. CLAIM_DIAGNOSIS (Junction: claim diagnoses with sequence)
-- ============================================================================

CREATE TABLE claim_diagnosis (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    claim_id            UUID NOT NULL REFERENCES claim(id) ON DELETE CASCADE,
    sequence            INTEGER NOT NULL,
    diagnosis_code_system VARCHAR(255),
    diagnosis_code      VARCHAR(30) NOT NULL,
    diagnosis_display   VARCHAR(500),
    type_code           VARCHAR(20),
    on_admission        BOOLEAN,
    package_code        VARCHAR(30)
);

-- ============================================================================
-- 4. CLAIM_PROCEDURE (Junction: claim procedures with sequence)
-- ============================================================================

CREATE TABLE claim_procedure (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    claim_id            UUID NOT NULL REFERENCES claim(id) ON DELETE CASCADE,
    sequence            INTEGER NOT NULL,
    type_code           VARCHAR(20),
    date                DATE,
    procedure_code_system VARCHAR(255),
    procedure_code      VARCHAR(30) NOT NULL,
    procedure_display   VARCHAR(500),
    udi                 VARCHAR(100)
);

-- ============================================================================
-- 5. CLAIM_ITEM (Line items)
-- ============================================================================

CREATE TABLE claim_item (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    claim_id            UUID NOT NULL REFERENCES claim(id) ON DELETE CASCADE,
    sequence            INTEGER NOT NULL,
    product_or_service_system VARCHAR(255),
    product_or_service_code VARCHAR(30) NOT NULL,
    product_or_service_display VARCHAR(500),
    serviced_date       DATE,
    serviced_period_start DATE,
    serviced_period_end   DATE,
    location_code       VARCHAR(5),
    quantity_value      DECIMAL(10,2),
    quantity_unit       VARCHAR(20),
    unit_price          DECIMAL(12,2),
    factor              DECIMAL(5,4),
    net_amount          DECIMAL(12,2),
    currency            VARCHAR(3) DEFAULT 'USD',
    revenue_code        VARCHAR(10),
    revenue_display     VARCHAR(255),
    body_site_code      VARCHAR(30),
    sub_site_code       VARCHAR(30),
    encounter_id        UUID REFERENCES encounter(id),
    note                TEXT
);

-- ============================================================================
-- 6. CLAIM_RESPONSE (Adjudication results)
-- ============================================================================

CREATE TABLE claim_response (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    claim_id            UUID NOT NULL REFERENCES claim(id),
    status              VARCHAR(20) NOT NULL,
    type_code           VARCHAR(30),
    use_code            VARCHAR(20),
    outcome             VARCHAR(20),
    disposition         TEXT,
    pre_auth_ref        VARCHAR(100),
    payment_type_code   VARCHAR(20),
    payment_adjustment  DECIMAL(12,2),
    payment_adjustment_reason VARCHAR(255),
    payment_amount      DECIMAL(12,2),
    payment_date        DATE,
    payment_identifier  VARCHAR(100),
    total_amount        DECIMAL(12,2),
    process_note        TEXT,
    communication_request TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 7. CLAIM_RESPONSE_ITEM (Adjudication details per item)
-- ============================================================================

CREATE TABLE claim_response_item (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    claim_response_id   UUID NOT NULL REFERENCES claim_response(id) ON DELETE CASCADE,
    item_sequence       INTEGER NOT NULL,
    adjudication_category VARCHAR(30),
    adjudication_amount DECIMAL(12,2),
    adjudication_value  DECIMAL(12,4),
    adjudication_reason VARCHAR(255),
    note                TEXT
);

-- ============================================================================
-- 8. EXPLANATION_OF_BENEFIT (EOBs)
-- ============================================================================

CREATE TABLE explanation_of_benefit (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    status              VARCHAR(20) NOT NULL,
    type_code           VARCHAR(30),
    use_code            VARCHAR(20),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    claim_id            UUID REFERENCES claim(id),
    claim_response_id   UUID REFERENCES claim_response(id),
    coverage_id         UUID REFERENCES coverage(id),
    insurer_org_id      UUID REFERENCES organization(id),
    provider_id         UUID REFERENCES practitioner(id),
    outcome             VARCHAR(20),
    disposition         TEXT,
    billable_period_start DATE,
    billable_period_end   DATE,
    total_submitted     DECIMAL(12,2),
    total_benefit       DECIMAL(12,2),
    total_patient_responsibility DECIMAL(12,2),
    total_payment       DECIMAL(12,2),
    payment_date        DATE,
    currency            VARCHAR(3) DEFAULT 'USD',
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 9. INVOICE (Direct billing)
-- ============================================================================

CREATE TABLE invoice (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    status              VARCHAR(20) NOT NULL,
    type_code           VARCHAR(30),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    encounter_id        UUID REFERENCES encounter(id),
    issuer_org_id       UUID REFERENCES organization(id),
    date                TIMESTAMPTZ DEFAULT NOW(),
    participant_id      UUID REFERENCES practitioner(id),
    total_net           DECIMAL(12,2),
    total_gross         DECIMAL(12,2),
    total_tax           DECIMAL(12,2),
    currency            VARCHAR(3) DEFAULT 'USD',
    payment_terms       TEXT,
    gstin               VARCHAR(20),
    gst_amount          DECIMAL(12,2),
    sac_code            VARCHAR(10),
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 10. INVOICE_LINE_ITEM (Invoice details)
-- ============================================================================

CREATE TABLE invoice_line_item (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id          UUID NOT NULL REFERENCES invoice(id) ON DELETE CASCADE,
    sequence            INTEGER NOT NULL,
    description         VARCHAR(500),
    service_code        VARCHAR(30),
    service_display     VARCHAR(255),
    quantity            DECIMAL(10,2),
    unit_price          DECIMAL(12,2),
    net_amount          DECIMAL(12,2),
    tax_amount          DECIMAL(12,2),
    gross_amount        DECIMAL(12,2),
    currency            VARCHAR(3) DEFAULT 'USD'
);

-- ============================================================================
-- 11. INDEXES for T2 billing tables
-- ============================================================================

-- Coverage indexes
CREATE INDEX idx_coverage_fhir ON coverage(fhir_id);
CREATE INDEX idx_coverage_patient ON coverage(patient_id);
CREATE INDEX idx_coverage_status ON coverage(status);
CREATE INDEX idx_coverage_payor ON coverage(payor_org_id) WHERE payor_org_id IS NOT NULL;

-- Claim indexes
CREATE INDEX idx_claim_fhir ON claim(fhir_id);
CREATE INDEX idx_claim_patient ON claim(patient_id);
CREATE INDEX idx_claim_status ON claim(status);
CREATE INDEX idx_claim_coverage ON claim(coverage_id) WHERE coverage_id IS NOT NULL;
CREATE INDEX idx_claim_encounter ON claim(encounter_id) WHERE encounter_id IS NOT NULL;
CREATE INDEX idx_claim_created ON claim(created_at DESC);

-- Claim diagnosis / procedure / item indexes
CREATE INDEX idx_claim_diag_claim ON claim_diagnosis(claim_id);
CREATE INDEX idx_claim_proc_claim ON claim_procedure(claim_id);
CREATE INDEX idx_claim_item_claim ON claim_item(claim_id);

-- Claim response indexes
CREATE INDEX idx_claim_resp_fhir ON claim_response(fhir_id);
CREATE INDEX idx_claim_resp_claim ON claim_response(claim_id);
CREATE INDEX idx_claim_resp_item_resp ON claim_response_item(claim_response_id);

-- EOB indexes
CREATE INDEX idx_eob_fhir ON explanation_of_benefit(fhir_id);
CREATE INDEX idx_eob_patient ON explanation_of_benefit(patient_id);
CREATE INDEX idx_eob_claim ON explanation_of_benefit(claim_id) WHERE claim_id IS NOT NULL;

-- Invoice indexes
CREATE INDEX idx_invoice_fhir ON invoice(fhir_id);
CREATE INDEX idx_invoice_patient ON invoice(patient_id);
CREATE INDEX idx_invoice_status ON invoice(status);
CREATE INDEX idx_invoice_line_invoice ON invoice_line_item(invoice_id);
