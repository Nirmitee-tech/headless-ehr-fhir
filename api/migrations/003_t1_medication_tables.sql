-- ============================================================================
-- T1 MEDICATION TABLES MIGRATION
-- Tables: medication, medication_ingredient, medication_request,
--         medication_administration, medication_dispense, medication_statement
-- References: patient, practitioner, encounter, condition, organization, location
-- ============================================================================

-- ============================================================================
-- 1. MEDICATION (Drug Catalog) (from 08_medications.sql)
-- ============================================================================

CREATE TABLE medication (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    -- Code
    code_system         VARCHAR(255),          -- http://www.nlm.nih.gov/research/umls/rxnorm | http://snomed.info/sct
    code_value          VARCHAR(30) NOT NULL,   -- RxNorm CUI or SNOMED code
    code_display        VARCHAR(500) NOT NULL,  -- Display name
    -- Status
    status              VARCHAR(20) NOT NULL DEFAULT 'active',  -- active | inactive | entered-in-error
    -- Form
    form_code           VARCHAR(30),            -- tablet | capsule | solution | injection | cream | patch | inhaler | suppository | drops | powder | spray | lozenge | suspension
    form_display        VARCHAR(255),
    -- Strength / Amount
    amount_numerator    DECIMAL(12,4),
    amount_numerator_unit VARCHAR(30),
    amount_denominator  DECIMAL(12,4),
    amount_denominator_unit VARCHAR(30),
    -- Schedule / Classification
    schedule            VARCHAR(10),            -- US DEA schedule: I | II | III | IV | V | OTC
    is_brand            BOOLEAN DEFAULT FALSE,
    is_over_the_counter BOOLEAN DEFAULT FALSE,
    -- Manufacturer
    manufacturer_id     UUID REFERENCES organization(id),
    manufacturer_name   VARCHAR(255),
    -- Batch
    lot_number          VARCHAR(50),
    expiration_date     DATE,
    -- NDC / Identifiers
    ndc_code            VARCHAR(20),            -- US: National Drug Code
    gtin_code           VARCHAR(20),            -- Global Trade Item Number
    -- India Specific
    dpco_scheduled      BOOLEAN DEFAULT FALSE,  -- India: Drug Price Control Order
    cdsco_approval      VARCHAR(50),            -- India: CDSCO approval number
    -- Flags
    is_narcotic         BOOLEAN DEFAULT FALSE,
    is_antibiotic       BOOLEAN DEFAULT FALSE,
    is_high_alert       BOOLEAN DEFAULT FALSE,
    requires_reconstitution BOOLEAN DEFAULT FALSE,
    -- Description
    description         TEXT,
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 2. MEDICATION INGREDIENT
-- ============================================================================

CREATE TABLE medication_ingredient (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    medication_id       UUID NOT NULL REFERENCES medication(id) ON DELETE CASCADE,
    -- Item (active ingredient)
    item_code           VARCHAR(30),
    item_display        VARCHAR(255) NOT NULL,
    item_system         VARCHAR(255),
    -- Strength
    strength_numerator      DECIMAL(12,4),
    strength_numerator_unit VARCHAR(30),
    strength_denominator    DECIMAL(12,4),
    strength_denominator_unit VARCHAR(30),
    -- Flags
    is_active           BOOLEAN DEFAULT TRUE
);

-- ============================================================================
-- 3. MEDICATION REQUEST (Prescription Orders)
-- ============================================================================

CREATE TABLE medication_request (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    -- Status & Intent
    status              VARCHAR(30) NOT NULL DEFAULT 'draft',  -- active | on-hold | cancelled | completed | entered-in-error | stopped | draft | unknown
    status_reason_code  VARCHAR(30),
    status_reason_display VARCHAR(255),
    intent              VARCHAR(30) NOT NULL DEFAULT 'order',  -- proposal | plan | order | original-order | reflex-order | filler-order | instance-order | option
    -- Category
    category_code       VARCHAR(30),            -- inpatient | outpatient | community | discharge
    category_display    VARCHAR(100),
    -- Priority
    priority            VARCHAR(20),            -- routine | urgent | asap | stat
    -- Medication
    medication_id       UUID NOT NULL REFERENCES medication(id),
    -- Subject & Context
    patient_id          UUID NOT NULL REFERENCES patient(id),
    encounter_id        UUID REFERENCES encounter(id),
    -- Requester / Performers
    requester_id        UUID NOT NULL REFERENCES practitioner(id),
    performer_id        UUID REFERENCES practitioner(id),
    recorder_id         UUID REFERENCES practitioner(id),
    -- Reason
    reason_code         VARCHAR(30),
    reason_display      VARCHAR(255),
    reason_condition_id UUID REFERENCES condition(id),
    -- Dosage Instruction
    dosage_text         TEXT,
    dosage_timing_code  VARCHAR(30),            -- QD | BID | TID | QID | Q4H | Q6H | Q8H | Q12H | HS | PRN | ONCE | STAT
    dosage_timing_display VARCHAR(100),
    dosage_route_code   VARCHAR(30),            -- PO | IV | IM | SC | SL | TOP | INH | PR | TD | OPTH | OT | NAS | NGTUBE | PEG
    dosage_route_display VARCHAR(100),
    dosage_site_code    VARCHAR(30),
    dosage_site_display VARCHAR(100),
    dosage_method_code  VARCHAR(30),
    dosage_method_display VARCHAR(100),
    dose_quantity       DECIMAL(12,4),
    dose_unit           VARCHAR(30),
    max_dose_per_period DECIMAL(12,4),
    max_dose_per_period_unit VARCHAR(30),
    rate_quantity       DECIMAL(12,4),
    rate_unit           VARCHAR(30),
    -- As Needed
    as_needed           BOOLEAN DEFAULT FALSE,
    as_needed_code      VARCHAR(30),
    as_needed_display   VARCHAR(100),
    -- Dispense Request
    quantity_value      DECIMAL(12,4),
    quantity_unit       VARCHAR(30),
    days_supply         INTEGER,
    refills_allowed     INTEGER DEFAULT 0,
    -- Validity Period
    validity_start      TIMESTAMPTZ,
    validity_end        TIMESTAMPTZ,
    -- Substitution
    substitution_allowed BOOLEAN DEFAULT TRUE,
    substitution_reason VARCHAR(255),
    -- Dates
    authored_on         TIMESTAMPTZ DEFAULT NOW(),
    -- Prior Authorization
    prior_auth_number   VARCHAR(50),
    -- e-Prescription
    erx_reference       VARCHAR(100),           -- e-prescription reference number
    -- India Specific
    abdm_prescription_id VARCHAR(50),
    -- Note
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 4. MEDICATION ADMINISTRATION (MAR - Medication Administration Record)
-- ============================================================================

CREATE TABLE medication_administration (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    -- Status
    status              VARCHAR(30) NOT NULL DEFAULT 'in-progress',  -- in-progress | not-done | on-hold | completed | entered-in-error | stopped | unknown
    status_reason_code  VARCHAR(30),
    status_reason_display VARCHAR(255),
    -- Category
    category_code       VARCHAR(30),            -- inpatient | outpatient | community
    category_display    VARCHAR(100),
    -- Medication
    medication_id       UUID NOT NULL REFERENCES medication(id),
    -- Subject & Context
    patient_id          UUID NOT NULL REFERENCES patient(id),
    encounter_id        UUID REFERENCES encounter(id),
    -- Request Reference
    medication_request_id UUID REFERENCES medication_request(id),
    -- Performer
    performer_id        UUID REFERENCES practitioner(id),
    performer_role_code VARCHAR(30),            -- performer | verifier | witness
    performer_role_display VARCHAR(100),
    -- Timing
    effective_datetime  TIMESTAMPTZ,
    effective_start     TIMESTAMPTZ,
    effective_end       TIMESTAMPTZ,
    -- Reason
    reason_code         VARCHAR(30),
    reason_display      VARCHAR(255),
    reason_condition_id UUID REFERENCES condition(id),
    -- Dosage
    dosage_text         TEXT,
    dosage_route_code   VARCHAR(30),
    dosage_route_display VARCHAR(100),
    dosage_site_code    VARCHAR(30),
    dosage_site_display VARCHAR(100),
    dosage_method_code  VARCHAR(30),
    dosage_method_display VARCHAR(100),
    dose_quantity       DECIMAL(12,4),
    dose_unit           VARCHAR(30),
    rate_quantity       DECIMAL(12,4),
    rate_unit           VARCHAR(30),
    -- Note
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 5. MEDICATION DISPENSE (Pharmacy)
-- ============================================================================

CREATE TABLE medication_dispense (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    -- Status
    status              VARCHAR(30) NOT NULL DEFAULT 'preparation',  -- preparation | in-progress | cancelled | on-hold | completed | entered-in-error | stopped | declined | unknown
    status_reason_code  VARCHAR(30),
    status_reason_display VARCHAR(255),
    -- Category
    category_code       VARCHAR(30),
    category_display    VARCHAR(100),
    -- Medication
    medication_id       UUID NOT NULL REFERENCES medication(id),
    -- Subject & Context
    patient_id          UUID NOT NULL REFERENCES patient(id),
    encounter_id        UUID REFERENCES encounter(id),
    -- Request Reference
    medication_request_id UUID REFERENCES medication_request(id),
    -- Performers
    performer_id        UUID REFERENCES practitioner(id),
    location_id         UUID REFERENCES location(id),
    -- Quantity
    quantity_value      DECIMAL(12,4),
    quantity_unit       VARCHAR(30),
    days_supply         INTEGER,
    -- Dates
    when_prepared       TIMESTAMPTZ,
    when_handed_over    TIMESTAMPTZ,
    -- Destination / Receiver
    destination_id      UUID REFERENCES location(id),
    receiver_id         UUID REFERENCES practitioner(id),
    -- Substitution
    was_substituted     BOOLEAN DEFAULT FALSE,
    substitution_type_code VARCHAR(30),
    substitution_reason VARCHAR(255),
    -- Note
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 6. MEDICATION STATEMENT (Patient-Reported Medications)
-- ============================================================================

CREATE TABLE medication_statement (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    -- Status
    status              VARCHAR(30) NOT NULL DEFAULT 'active',  -- active | completed | entered-in-error | intended | stopped | on-hold | unknown | not-taken
    status_reason_code  VARCHAR(30),
    status_reason_display VARCHAR(255),
    -- Category
    category_code       VARCHAR(30),            -- inpatient | outpatient | community | patientspecified
    category_display    VARCHAR(100),
    -- Medication
    medication_code     VARCHAR(30),
    medication_display  VARCHAR(500),
    medication_id       UUID REFERENCES medication(id),
    -- Subject & Context
    patient_id          UUID NOT NULL REFERENCES patient(id),
    encounter_id        UUID REFERENCES encounter(id),
    -- Source (who reported)
    information_source_id UUID REFERENCES practitioner(id),
    -- Dates
    effective_datetime  TIMESTAMPTZ,
    effective_start     TIMESTAMPTZ,
    effective_end       TIMESTAMPTZ,
    date_asserted       TIMESTAMPTZ DEFAULT NOW(),
    -- Reason
    reason_code         VARCHAR(30),
    reason_display      VARCHAR(255),
    -- Dosage
    dosage_text         TEXT,
    dosage_route_code   VARCHAR(30),
    dosage_route_display VARCHAR(100),
    dose_quantity       DECIMAL(12,4),
    dose_unit           VARCHAR(30),
    dosage_timing_code  VARCHAR(30),
    dosage_timing_display VARCHAR(100),
    -- Note
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 7. INDEXES for Medication tables
-- ============================================================================

-- Medication indexes
CREATE INDEX idx_medication_fhir ON medication(fhir_id);
CREATE INDEX idx_medication_code ON medication(code_value);
CREATE INDEX idx_medication_status ON medication(status);
CREATE INDEX idx_medication_ndc ON medication(ndc_code) WHERE ndc_code IS NOT NULL;
CREATE INDEX idx_medication_schedule ON medication(schedule) WHERE schedule IS NOT NULL;

-- Medication Ingredient indexes
CREATE INDEX idx_med_ingredient_medication ON medication_ingredient(medication_id);

-- Medication Request indexes
CREATE INDEX idx_med_request_fhir ON medication_request(fhir_id);
CREATE INDEX idx_med_request_patient ON medication_request(patient_id);
CREATE INDEX idx_med_request_encounter ON medication_request(encounter_id) WHERE encounter_id IS NOT NULL;
CREATE INDEX idx_med_request_medication ON medication_request(medication_id);
CREATE INDEX idx_med_request_requester ON medication_request(requester_id);
CREATE INDEX idx_med_request_status ON medication_request(status);
CREATE INDEX idx_med_request_intent ON medication_request(intent);
CREATE INDEX idx_med_request_authored ON medication_request(authored_on DESC);
CREATE INDEX idx_med_request_patient_status ON medication_request(patient_id, status);

-- Medication Administration indexes
CREATE INDEX idx_med_admin_fhir ON medication_administration(fhir_id);
CREATE INDEX idx_med_admin_patient ON medication_administration(patient_id);
CREATE INDEX idx_med_admin_encounter ON medication_administration(encounter_id) WHERE encounter_id IS NOT NULL;
CREATE INDEX idx_med_admin_medication ON medication_administration(medication_id);
CREATE INDEX idx_med_admin_status ON medication_administration(status);
CREATE INDEX idx_med_admin_effective ON medication_administration(effective_datetime DESC) WHERE effective_datetime IS NOT NULL;
CREATE INDEX idx_med_admin_request ON medication_administration(medication_request_id) WHERE medication_request_id IS NOT NULL;
CREATE INDEX idx_med_admin_patient_status ON medication_administration(patient_id, status);

-- Medication Dispense indexes
CREATE INDEX idx_med_dispense_fhir ON medication_dispense(fhir_id);
CREATE INDEX idx_med_dispense_patient ON medication_dispense(patient_id);
CREATE INDEX idx_med_dispense_encounter ON medication_dispense(encounter_id) WHERE encounter_id IS NOT NULL;
CREATE INDEX idx_med_dispense_medication ON medication_dispense(medication_id);
CREATE INDEX idx_med_dispense_status ON medication_dispense(status);
CREATE INDEX idx_med_dispense_request ON medication_dispense(medication_request_id) WHERE medication_request_id IS NOT NULL;
CREATE INDEX idx_med_dispense_handed_over ON medication_dispense(when_handed_over DESC) WHERE when_handed_over IS NOT NULL;

-- Medication Statement indexes
CREATE INDEX idx_med_statement_fhir ON medication_statement(fhir_id);
CREATE INDEX idx_med_statement_patient ON medication_statement(patient_id);
CREATE INDEX idx_med_statement_encounter ON medication_statement(encounter_id) WHERE encounter_id IS NOT NULL;
CREATE INDEX idx_med_statement_status ON medication_statement(status);
CREATE INDEX idx_med_statement_effective ON medication_statement(effective_datetime DESC) WHERE effective_datetime IS NOT NULL;
CREATE INDEX idx_med_statement_patient_status ON medication_statement(patient_id, status);
